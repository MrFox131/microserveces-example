package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/log"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"microserveces-example/account"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const dbSource = "postgresql://postgres:postgres@postgres-service:5432/gokit?sslmode=disable"

const amqpSource = "amqp://guest:guest@rabbitmq:5672/"

func failOnError(err error, logger log.Logger) {
	if err != nil {
		level.Error(logger).Log("error", err)
		os.Exit(-1)
	}
}

func main() {
	var httpAddr = flag.String("http", ":8080", "http listen address")
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger, "service", "account", "time:", log.DefaultTimestamp, "caller: ", log.DefaultCaller)
	}

	level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service stopped")

	amqpServer, err := amqp.Dial(amqpSource)
	if err != nil {
		level.Error(logger).Log("error", "unable to connect to amqp server", "reason", err)
		os.Exit(-1)
	}

	var db *sql.DB
	{
		var err error

		db, err = sql.Open("postgres", dbSource)
		failOnError(err, logger)
	}
	flag.Parse()

	ctx := context.Background()
	var srv account.Service
	{
		repository := account.NewRepo(db, logger)
		srv = account.NewService(repository, logger)
	}

	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	endpoints := account.MakeEndpoints(srv)

	go func() {
		fmt.Println("listening on port", *httpAddr)
		handler := account.NewHTTPServer(ctx, endpoints)
		errs <- http.ListenAndServe(*httpAddr, handler)
	}()

	{
		fmt.Println("starting amqp listener")
		subs := account.NewAMQPSubscribers(ctx, endpoints)

		go func() {
			ch, err := amqpServer.Channel()
			err = ch.ExchangeDeclare(
				"external_exchange", // name
				"direct",            // type
				true,                // durable
				false,               // auto-deleted
				false,               // internal
				false,               // no-wait
				nil,                 // arguments
			)
			failOnError(err, logger)
			consumer := subs.CrateUserSubscriber.ServeDelivery(ch)
			queue, err := ch.QueueDeclare(
				"test_queue",
				false,
				false,
				false,
				false,
				nil,
			)
			failOnError(err, logger)

			err = ch.QueueBind(
				queue.Name,
				"user.create",
				"external_exchange",
				false,
				nil,
			)
			failOnError(err, logger)

			msgs, err := ch.Consume(
				"test_queue",
				"",
				true,
				false,
				false,
				false,
				nil,
			)
			failOnError(err, logger)

			for msg := range msgs {
				consumer(&msg)
			}
		}()
		go func() {
			ch, err := amqpServer.Channel()
			err = ch.ExchangeDeclare(
				"external_exchange", // name
				"direct",            // type
				true,                // durable
				false,               // auto-deleted
				false,               // internal
				false,               // no-wait
				nil,                 // arguments
			)
			failOnError(err, logger)

			consumer := subs.GetUserSubscriber.ServeDelivery(ch)
			queue, err := ch.QueueDeclare(
				"getUserQueue",
				false,
				false,
				false,
				false,
				nil,
			)
			failOnError(err, logger)

			err = ch.QueueBind(
				queue.Name,
				"user.get",
				"external_exchange",
				false,
				nil,
			)
			failOnError(err, logger)

			msgs, err := ch.Consume(
				queue.Name,
				"",
				true,
				false,
				false,
				false,
				nil,
			)
			failOnError(err, logger)

			for msg := range msgs {
				consumer(&msg)
			}
		}()
	}

	level.Error(logger).Log("exit", <-errs)
}
