package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"github.com/galifornia/go-micro-logger/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	WEB_PORT  = "80"
	RPC_PORT  = "5001"
	GRPC_PORT = "50001"
	MONGO_URL = "mongodb://mongo:27017"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	// register & start rpc server
	err = rpc.Register(new(RPCServer))
	if err != nil {
		panic(err)
	}
	go app.rpcListen()

	// start web server
	app.serve()
}

func (app *Config) serve() {
	log.Println("Starting web server")
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", WEB_PORT),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic("Unable to start web server")
	}
}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on Port", RPC_PORT)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", RPC_PORT))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)
	}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(MONGO_URL)

	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	conn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Connected to Mongo!!")
	return conn, nil
}
