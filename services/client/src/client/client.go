package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/entities"
	filehandler "github.com/7574-sistemas-distribuidos/tp-nivelador/src/file-handler"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/signals"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost    string
	ServerPort    string
	AgencyId      string
	InputFilePath string
	OutputDir     string
	BatchSize     string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func checkError(ctx context.Context, action string, err error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return nil
	}
	logger.Error(action, logger.Fail, "err", err)
	return err
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func sendBets(ctx context.Context, client *Client, records [][]string) error {
	bets, err := entities.BetsFromRecords(records)
	if err := checkError(ctx, "bets-from-records", err); err != nil {
		return err
	}

	betPacket, err := protocol.GenerateBetPacket(bets, client.config.AgencyId)
	if err := checkError(ctx, "generate-bet-packet", err); err != nil {
		return err
	}

	if err := checkError(ctx, "send-packet", safe_socket.SendAll(client.conn, betPacket)); err != nil {
		return err
	}

	ack, err := safe_socket.RecvAll(client.conn, 1)
	if err := checkError(ctx, "recv-ack", err); err != nil {
		return err
	}
	if len(ack) != 1 || ack[0] != byte(protocol.AckBetMessage) {
		return checkError(ctx, "recv-ack", fmt.Errorf("expected AckBetMessage, received %v", ack))
	}
	return nil
}

func receiveWinnersAndStoreInCSV(ctx context.Context, client *Client) error {
	const mainAction = "receive-winners-and-store-in-csv"
	payloadLengthBytes, err := safe_socket.RecvAll(client.conn, 4)
	if err := checkError(ctx, "recv-winners-bets-amount", err); err != nil {
		return err
	}

	payloadLength, err := binary.BigEndian.Uint32(payloadLengthBytes), nil

	logger.Info(mainAction, logger.InProgress, "payload-length", payloadLength)
	payload, err := safe_socket.RecvAll(client.conn, int(payloadLength))
	if err := checkError(ctx, "recv-winners-bets-payload", err); err != nil {
		return err
	}
	winners, err := protocol.UnpackWinners(payload)
	if err := checkError(ctx, "unpack-winners-bets", err); err != nil {
		return err
	}
	var records [][]string
	for _, winner := range winners {
		record := entities.BetToRecord(winner)
		records = append(records, record)
	}
	outputFilePath := client.config.OutputDir

	return checkError(ctx, "create-csv-writer", filehandler.WriteCSVFile(outputFilePath, records))
}

func (client *Client) Run() error {
	const mainAction = "run-client"

	ctx, stop := signals.SetupSignalHandler(client.conn)
	defer stop()
	defer func() {
		if client.conn != nil {
			_ = client.conn.Close()
		}
	}()

	packetSize, _ := strconv.Atoi(client.config.BatchSize)
	thereArePacketsToSend := true
	reader, err := filehandler.NewCSVReader(client.config.InputFilePath)
	if err := checkError(ctx, "read-csv-file", err); err != nil {
		return err
	}
	defer reader.Close()

	for thereArePacketsToSend {
		if ctx.Err() != nil {
			return nil
		}

		records, hasMoreRecords, err := reader.ReadBatch(packetSize)
		if err := checkError(ctx, "read-batch", err); err != nil {
			return err
		}

		if len(records) > 0 {
			sendBetsErr := sendBets(ctx, client, records)
			if err := checkError(ctx, "send-bets", sendBetsErr); err != nil {
				return err
			}
		}
		thereArePacketsToSend = hasMoreRecords
	}

	allBetsSentPacket, err := protocol.GenerateAllBetsSentPacket(client.config.AgencyId)
	if err := checkError(ctx, "generate-all-bets-sent-packet", err); err != nil {
		return err
	}

	if err := checkError(ctx, "send-all-bets-sent-packet", safe_socket.SendAll(client.conn, allBetsSentPacket)); err != nil {
		return err
	}
	response, err := safe_socket.RecvAll(client.conn, 1)
	if err := checkError(ctx, "recv-response", err); err != nil {
		return err
	}

	isWinnerResponse, err := protocol.UnpackMessageType(response)
	if err := checkError(ctx, "unpack-response", err); err != nil {
		return err
	}

	if !isWinnerResponse {
		return checkError(ctx, "unpack-response", fmt.Errorf("expected WinnersMessage, received unexpected message type"))
	}
	return receiveWinnersAndStoreInCSV(ctx, client)
}
