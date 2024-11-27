# Peril: A Risk Game Clone
Welcome to Peril, a clone of the classic Risk game, developed using Go and RabbitMQ.

## Acknowledgements

This project is part of the Pub/Sub Architecture course by [Boot.dev](boot.dev).   
I would like to thank Boot.dev for their comprehensive and insightful material. 



## Introduction
 Peril is a strategic board game where players compete to dominate the world by controlling armies and conquering territories. This project implements the core mechanics of Risk using Go for the game logic and RabbitMQ for the pub/sub messaging system.

## Features
- Multiplayer gameplay
- Territory control and management
- Army deployment and battles
- Turn-based actions
- Real-time updates using RabbitMQ

## Installation
- Clone the repository:

```sh
  Copy code
  git clone https://github.com/yourusername/peril.git
  cd peril
```

- Install Go dependencies:

``` sh
  Copy code
  go mod tidy
```


## Usage
- Run the RabbitMQ server:

``` sh
  ./rabbit.sh
```
- Run the Peril server:

```sh
go run ./server/main.go
``` 
- Connect to the game via the client for each player:
```sh
go run ./client/main.go
``` 
