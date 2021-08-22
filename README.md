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
// TODO
```

## Production reccomendations
1. Run this behind a reverse proxy or in a cluster like Kubernetes to get HTTPS
2. Use secure tokens and only connect over WSS