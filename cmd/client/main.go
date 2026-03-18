package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/FIFSAK/TMS/pkg/pb/shipment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewShipmentServiceClient(conn)
	fmt.Printf("Connected to %s\n\n", *addr)

	// 1. Create a shipment
	fmt.Println("=== CreateShipment ===")
	createResp, err := client.CreateShipment(ctx, &pb.CreateShipmentRequest{
		ReferenceNumber: "SHP-2026-001",
		Origin:          "New York, NY",
		Destination:     "Los Angeles, CA",
		DriverName:      "John Smith",
		UnitNumber:      "TRUCK-42",
		ShipmentAmount:  5000.00,
		DriverRevenue:   1500.00,
	})
	if err != nil {
		log.Fatalf("CreateShipment: %v", err)
	}
	shipment := createResp.GetShipment()
	fmt.Printf("Created shipment: id=%s ref=%s status=%s\n\n",
		shipment.GetId(), shipment.GetReferenceNumber(), shipment.GetStatus())

	shipmentID := shipment.GetId()

	// 2. Get the shipment
	fmt.Println("=== GetShipment ===")
	getResp, err := client.GetShipment(ctx, &pb.GetShipmentRequest{Id: shipmentID})
	if err != nil {
		log.Fatalf("GetShipment: %v", err)
	}
	s := getResp.GetShipment()
	fmt.Printf("Got shipment: id=%s origin=%s dest=%s status=%s amount=%.2f\n\n",
		s.GetId(), s.GetOrigin(), s.GetDestination(), s.GetStatus(), s.GetShipmentAmount())

	// 3. Walk through the lifecycle: picked_up → in_transit → delivered
	transitions := []struct {
		status  pb.ShipmentStatus
		comment string
	}{
		{pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP, "Driver picked up the shipment"},
		{pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT, "Shipment is en route"},
		{pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED, "Shipment delivered to destination"},
	}

	for _, t := range transitions {
		fmt.Printf("=== AddShipmentEvent: %s ===\n", t.status)
		eventResp, err := client.AddShipmentEvent(ctx, &pb.AddShipmentEventRequest{
			ShipmentId: shipmentID,
			Status:     t.status,
			Comment:    t.comment,
		})
		if err != nil {
			log.Fatalf("AddShipmentEvent(%s): %v", t.status, err)
		}
		evt := eventResp.GetEvent()
		fmt.Printf("Event added: id=%s status=%s comment=%q\n\n", evt.GetId(), evt.GetStatus(), evt.GetComment())
	}

	// 4. Try an invalid transition (delivered → picked_up)
	fmt.Println("=== AddShipmentEvent: invalid transition (delivered → picked_up) ===")
	_, err = client.AddShipmentEvent(ctx, &pb.AddShipmentEventRequest{
		ShipmentId: shipmentID,
		Status:     pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP,
		Comment:    "This should fail",
	})
	if err != nil {
		fmt.Printf("Expected error: %v\n\n", err)
	} else {
		fmt.Println("ERROR: should have been rejected!")
		os.Exit(1)
	}

	// 5. Get full event history
	fmt.Println("=== GetShipmentEvents ===")
	eventsResp, err := client.GetShipmentEvents(ctx, &pb.GetShipmentEventsRequest{ShipmentId: shipmentID})
	if err != nil {
		log.Fatalf("GetShipmentEvents: %v", err)
	}
	for i, e := range eventsResp.GetEvents() {
		fmt.Printf("  [%d] status=%s comment=%q time=%s\n",
			i+1, e.GetStatus(), e.GetComment(), e.GetCreatedAt().AsTime().Format(time.RFC3339))
	}
	fmt.Println()

	// 6. List all shipments
	fmt.Println("=== ListShipments ===")
	listResp, err := client.ListShipments(ctx, &pb.ListShipmentsRequest{})
	if err != nil {
		log.Fatalf("ListShipments: %v", err)
	}
	for _, sh := range listResp.GetShipments() {
		fmt.Printf("  id=%s ref=%s status=%s\n", sh.GetId(), sh.GetReferenceNumber(), sh.GetStatus())
	}

	fmt.Println("\nAll tests passed!")
}
