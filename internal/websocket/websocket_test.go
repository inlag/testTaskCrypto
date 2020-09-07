package websocket

import (
	"net/http"
	"testing"
)

var (
	validURL, validHeader = (&WebSocket{}).DefaultHost()
)

func TestWebSocket_Connection(t *testing.T) {

	type args struct {
		url    string
		header http.Header
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestWithEmptyUrl",
			args: args{
				url:    "",
				header: http.Header{},
			},
			wantErr: true,
		},
		{
			name: "TestWithEmptyHeader",
			args: args{
				url:    validURL,
				header: nil,
			},
			wantErr: true,
		},
		{
			name: "TestWithValidParameters",
			args: args{
				url:    validURL,
				header: validHeader,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WebSocket{}
			if err := w.Connection(tt.args.url, tt.args.header); (err != nil) != tt.wantErr {
				t.Errorf("Connection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebSocket_Subscribe(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			url    string
			header http.Header
		}
		wantErr bool
	}{
		{
			name: "TestWithEmptyWssConnection",
			args: struct {
				url    string
				header http.Header
			}{
				url:    "",
				header: nil,
			},
			wantErr: true,
		},
		{
			name: "TestWithValidWssConnection",
			args: struct {
				url    string
				header http.Header
			}{
				url:    validURL,
				header: validHeader,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WebSocket{}
			_ = w.Connection(tt.args.url, tt.args.header)
			if err := w.Subscribe(); (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebSocket_Next(t *testing.T) {
	tests := []struct {
		name    string
		wss     func() *WebSocket
		wantErr bool
	}{
		{
			name: "TestWithValidConnection",
			wss: func() *WebSocket {
				w := &WebSocket{}
				_ = w.ConnectionAndSubscribe()
				return w
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := tt.wss()
			_, err := ws.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("Next() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
