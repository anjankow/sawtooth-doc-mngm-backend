// based on https://www.hyperledger.org/blog/2019/02/19/hyperledger-sawtooth-events-in-go-2
package events

import (
	"errors"
	"fmt"

	"github.com/hyperledger/sawtooth-sdk-go/messaging"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/client_event_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/events_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/validator_pb2"
	"github.com/pebbe/zmq4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type EventListener struct {
	log          *zap.Logger
	connection   messaging.Connection
	validatorUrl string
	closerFunc   []func() error
}

func NewEventListener(logger *zap.Logger, validatorHostname string) EventListener {
	validatorUrl := fmt.Sprint("tcp://", validatorHostname, ":4004")
	return EventListener{
		log:          logger,
		validatorUrl: validatorUrl,
	}
}

func (e EventListener) Start() error {
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

	e.subscribeToEvent("proposal_accepted")

	return nil
}

func (e EventListener) Stop() {
	for _, close := range e.closerFunc {
		if err := close(); err != nil {
			e.log.Error("event listener closer failed: " + err.Error())
		}
	}
	e.connection.Close()
}

func (e EventListener) listenLoop() error {
	for {
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
			switch event.EventType {
			case "proposal_accepted":
			default:
				e.log.Warn("handler missing for the event: " + event.EventType)
			}
		}
	}
}

func (e EventListener) subscribeToEvent(eventType string) (err error) {

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

	return nil
}
