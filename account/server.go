package account

import (
	"context"
	"net/http"

	amqptransport "github.com/go-kit/kit/transport/amqp"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

type AmqpHandlers struct {
	CrateUserSubscriber *amqptransport.Subscriber
	GetUserSubscriber   *amqptransport.Subscriber
}

type AmqpSubscriber amqptransport.Subscriber

func NewHTTPServer(_ context.Context, endpoints Endpoints) http.Handler {
	r := mux.NewRouter()
	r.Use(commonMiddleware)

	r.Methods("POST").Path("/user").Handler(httptransport.NewServer(
		endpoints.CreateUser,
		decodeUserReq,
		encodeResponse,
	))
	r.Methods("GET").Path("/user/{ID}").Handler(httptransport.NewServer(
		endpoints.GetUser,
		decodeEmailReq,
		encodeResponse,
	))

	return r
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func NewAMQPSubscribers(_ context.Context, e Endpoints) AmqpHandlers {
	createUserSubscriber := amqptransport.NewSubscriber(e.CreateUser, decodeUserAMQPReq, encodeAMQPResponse, amqptransport.SubscriberErrorEncoder(amqptransport.ReplyErrorEncoder))
	getUserSubscriber := amqptransport.NewSubscriber(e.GetUser, decodeEmailAMQPReq, encodeAMQPResponse, amqptransport.SubscriberErrorEncoder(amqptransport.ReplyErrorEncoder))
	return AmqpHandlers{
		createUserSubscriber,
		getUserSubscriber,
	}
}

//func (s *AmqpSubscriber) StartConsume(ch *amqp.Channel, queueName string, errs chan error) {
//	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
//	if err != nil{
//		errs <- err
//		return
//	}
//	for msg := range msgs {
//		s.Consume()
//	}
//
//}
