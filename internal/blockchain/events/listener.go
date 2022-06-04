// based on https://www.hyperledger.org/blog/2019/02/19/hyperledger-sawtooth-events-in-go-2
package events

import (
	"errors"
	"fmt"
	"sync"

	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/client_event_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/events_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	"github.com/pebbe/zmq4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type EventListener struct {
	log           *zap.Logger
	connection    messaging.Connection
	validatorUrl  string
	closerFunc    []func() error
	handlers      map[string](func(data []byte) error)
	stopListening chan bool
	wg            *sync.WaitGroup
}

func NewEventListener(logger *zap.Logger, validatorHostname string) *EventListener {
	validatorUrl := fmt.Sprint("tcp://", validatorHostname, ":4004")
	return &EventListener{
		log:          logger,
		validatorUrl: validatorUrl,
		handlers:     make(map[string]func(data []byte) error),
		wg:           &sync.WaitGroup{},
	}
}

func (e *EventListener) Start() error {
	zmqContext, err := zmq4.NewContext()
	if err != nil {
		return err
	}

	zmqConnection, err := messaging.NewConnection(
		zmqContext,
		zmq4.DEALER,
		e.validatorUrl,
		false,
	)
	if err != nil {
		return err
	}
	e.connection = zmqConnection

	for eventType, _ := range e.handlers {
		if err := e.subscribeToEvent(eventType); err != nil {
			e.log.Error("error when subscribing to event " + eventType + ": " + err.Error())
		}
	}

	e.stopListening = make(chan bool)
	go e.listenLoop(e.stopListening)

	return nil
}

func (e EventListener) Stop() error {
	e.stopListening <- true
	var allErr error
	for _, close := range e.closerFunc {
		if err := close(); err != nil {
			allErr = multierr.Append(allErr, err)
		}
	}
	e.connection.Close()
	e.log.Info("waiting for all the event handlers to finish...")
	e.wg.Wait()
	e.log.Info("event listener handlers finished")

	return allErr
}

func (e *EventListener) listenLoop(stop chan bool) error {
	e.log.Info("start listening on blockchain events")

	for {
		select {
		case <-stop:
			return nil
		default:

			// Wait for a message on connection
			_, message, err := e.connection.RecvMsg()
			if err != nil {
				return err
			}
			// Check if received is a client event message
			if message.MessageType !=
				validator_pb2.Message_CLIENT_EVENTS {
				return errors.New("Received a message not requested for")
			}
			event_list := events_pb2.EventList{}
			err = proto.Unmarshal(message.Content, &event_list)
			if err != nil {
				return err
			}
			// Received following events from validator
			for _, event := range event_list.Events {
				// handle event here
				e.log.Info("event received: " + event.EventType)

				handler, ok := e.handlers[event.EventType]
				if !ok {
					e.log.Warn("handler missing for the event: " + event.EventType)
				}

				e.wg.Add(1)
				go func(event *events_pb2.Event) {
					defer e.wg.Done()

					if err := handler(event.GetData()); err != nil {
						e.log.Error("error when handling the event: " + err.Error())
					}
				}(event)
			}
		}
	}
}

func (e *EventListener) SetHandler(eventType string, handler func(data []byte) error) error {
	e.handlers[eventType] = handler
	return nil
}

func (e *EventListener) subscribeToEvent(eventType string) (err error) {

	subs := events_pb2.EventSubscription{
		EventType: eventType,
	}
	request := client_event_pb2.ClientEventsSubscribeRequest{
		Subscriptions: []*events_pb2.EventSubscription{
			&subs,
		},
	}

	serializedReq, err := proto.Marshal(&request)
	if err != nil {
		return
	}
	// Send the subscription request, get a correlation id
	// from the SDK
	corrId, err := e.connection.SendNewMsg(
		validator_pb2.Message_CLIENT_EVENTS_SUBSCRIBE_REQUEST,
		serializedReq,
	)
	// Error requesting validator, optionally based on
	// error type may apply retry mechanism here
	if err != nil {
		return
	}
	e.log.Debug("waiting for receiving the subscription confirmation...")
	// Wait for subscription status, wait for response of
	// message with specific correlation id
	_, response, err := e.connection.RecvMsgWithId(corrId)
	if err != nil {
		return
	}

	// Deserialize received protobuf message as response
	// for subscription request
	subsResponse :=
		client_event_pb2.ClientEventsSubscribeResponse{}

	err = proto.Unmarshal(response.Content, &subsResponse)
	if err != nil {
		return
	}
	// Client subscription is not successful, optional
	// retries can be done later for subscription based on
	// response cause
	if subsResponse.Status !=
		client_event_pb2.ClientEventsSubscribeResponse_OK {
		return errors.New("client subscription failed, subscription status: " + subsResponse.String())
	}

	// Client event subscription is successful, remember to
	// unsubscribe when either not required anymore or
	// error occurs. Similar approach as followed for
	// subscribing events can be used here.
	unsubscribe := func() error {
		// Unsubscribe from events
		events_unsubscribe_request :=
			client_event_pb2.ClientEventsUnsubscribeRequest{}
		serialized_unsubscribe_request, err :=
			proto.Marshal(&events_unsubscribe_request)
		if err != nil {
			return err
		}

		corrId, err = e.connection.SendNewMsg(
			validator_pb2.Message_CLIENT_EVENTS_UNSUBSCRIBE_REQUEST,
			serialized_unsubscribe_request,
		)
		if err != nil {
			return err
		}
		// Wait for status
		_, unsubscribe_response, err :=
			e.connection.RecvMsgWithId(corrId)
		// Optional retries can be done depending on error status
		if err != nil {
			return err
		}
		events_unsubscribe_response := client_event_pb2.ClientEventsUnsubscribeResponse{}
		err = proto.Unmarshal(unsubscribe_response.Content,
			&events_unsubscribe_response)
		if err != nil {
			return err
		}
		// Optional retries can be done depending on error
		// status
		if events_unsubscribe_response.Status !=
			client_event_pb2.ClientEventsUnsubscribeResponse_OK {
			return errors.New("client couldn't unsubscribe successfully, status: " + events_unsubscribe_response.String())
		}

		return nil
	}

	e.closerFunc = append(e.closerFunc, unsubscribe)
	e.log.Info("successfully subscribed to event '" + eventType + "'")

	return nil
}
