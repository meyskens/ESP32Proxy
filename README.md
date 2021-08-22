# ESP32Proxy

The ESP32 is an excellent web server thanks to on board flash storage and decent wifi. However exposing it to the internet can be a challenge as TLS is quite hard to impossible, especially if you plan to depend on 4G connectivity (insert CGNAT rant here). 

This is where ESP32Proxy comes in handy, it is a server written in Go that can proxy several domains to an ESP32 over a websocket. The ESP32 connects to the server and authenthicates with a unique token, the server will from then on forward HTTP traffic of a given domain to the ESP over the websocket making the connection work on all kind of networks.

## Running the server
First of all you have to clone the project and edit `config.json`. You can set multiple domains or one. Make sure the token is secure as the holder of the token can host content on that domain!

Then start the program on a server you have exposed somewhere
```console
$ go run ./cmd/esp32proxy host # you can also change the port with -p
```

Congrats it is running! Now you need to connect the ESP32 to it over a websocket.

## ESP32 Code
```cpp
#include <Arduino.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <WebSocketsClient.h>

WebSocketsClient webSocket;


// Replace with your network credentials
const char* ssid = "";
const char* password = "";

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
  String request;
  String path;

    switch(type) {
        case WStype_DISCONNECTED:
            Serial.printf("[WSc] Disconnected!\n");
            break;
        case WStype_CONNECTED:
            Serial.printf("[WSc] Connected to url: %s\n", payload);
        case WStype_BIN:
          Serial.printf("[WSc] get binary length: %u\n", length);
          // parse HTTP body
          request = String((char*)payload);
          path = request.substring(request.indexOf("GET") + 4, request.indexOf("HTTP") - 1);

          if (path.equals("/owo")) {
            webSocket.sendTXT("HTTP/1.1 200 OK\r\nServer: ESP32\r\nX-Powered-By: ESP32Proxy\r\nConnection: close\r\nContent-Type: text/plain\r\n\r\nuwu");
            break;
          }

          webSocket.sendTXT("HTTP/1.1 200 OK\r\nServer: ESP32\r\nX-Powered-By: ESP32Proxy\r\nConnection: close\r\nContent-Type: text/html\r\n\r\nHello ESP");
          break;
        case WStype_ERROR:			
        case WStype_FRAGMENT_TEXT_START:
        case WStype_FRAGMENT_BIN_START:
        case WStype_FRAGMENT:
        case WStype_TEXT:
        case WStype_FRAGMENT_FIN:
            break;
    }

}

void setup(){
  Serial.begin(115200);

  // Connect to Wi-Fi
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(1000);
    Serial.println("Connecting to WiFi..");
  }

  Serial.println("Connected to Wifi, Connecting to server.");
  
  // server address, port and URL
  webSocket.begin("HOSTNAME", 80, "/proxy");
    // webSocket.beginSSL("HOSTNAME", 443, "/proxy"); // WSS

    // configure WS
    webSocket.onEvent(webSocketEvent);
  webSocket.setExtraHeaders("Token:TOKEN HERE"); // add token here
    webSocket.setReconnectInterval(5000);
}


void loop() {
  webSocket.loop();
}
```

## Production reccomendations
1. Run this behind a reverse proxy or in a cluster like Kubernetes to get HTTPS
2. Use secure tokens and only connect over WSS
