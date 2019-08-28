package main

import (
	"context"
	"fmt"
	"net"

	bs "github.com/bytecamp2019d/bustsurvivor/api/bustsurvivor"
	"github.com/bytecamp2019d/bustsurvivor/pkg/blackjack"
	"google.golang.org/grpc"
)

// Serve : Bind TCP With gRPC listener
func Serve() {
	srv := grpc.NewServer()
	bs.RegisterSurvivalServiceServer(srv, &Server{})
	address := ":8080"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on %s: %+v", address, err))
	}
	if err := srv.Serve(listener); err != nil {
		panic(fmt.Sprintf("gRPC serve failed: %+v", err))
	}
}

// Server is a bust survival server
type Server struct{}

// BustSurvival implement gRPC service
func (s *Server) BustSurvival(ctx context.Context, req *bs.BustSurvivalRequest) (*bs.BustSurvivalResponse, error) {
	resp := &bs.BustSurvivalResponse{}
	numerator, denominator, err := blackjack.Survival(req.CardsToPick, req.BustThreshold)
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Probability = &bs.Fraction{
			Numerator:   numerator,
			Denominator: denominator,
		}
	}
	return resp, nil
}

func main() {
	Serve()
	//1Serve1()
}
