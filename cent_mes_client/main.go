package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	addrString string = "http://192.168.0.100/echo"
)

func main() {
	ctx := context.Background()
	client := http.Client{}
	body := map[string]string{
		"mes": "bob",
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println(err)
	}
	req, err := http.NewRequestWithContext(
		ctx, "POST", addrString, bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
