package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"d-exclaimation.me/relax/config"
	"d-exclaimation.me/relax/lib/async"
	"d-exclaimation.me/relax/lib/f"
)

// KVPacket is a generic response from the KV store.
type KVPacket[Data any] struct {
	// The data itself.
	Result Data `json:"result"`
}

// KVPrimitive is a generic primitive type for the KV store.
type KVPrimitive interface {
	string
}

// KVCommand is a generic command to the KV store.
type KVCommand struct {
	Name string
	Args []any
}

const (
	get  = "GET"
	set  = "SET"
	incr = "INCR"
	mget = "MGET"
)

// Command is a generic command to the KV store.
func Command[Data any](name string, args ...any) async.Task[KVPacket[Data]] {
	return async.New(func() (KVPacket[Data], error) {
		command := make([]any, len(args)+1)
		command[0] = name
		for i, arg := range args {
			command[i+1] = arg
		}

		data := KVPacket[Data]{}
		body, err := json.Marshal(command)

		if err != nil {
			return data, err
		}

		req, err := http.NewRequest("POST", config.Env.KVURL(), bytes.NewBuffer(body))
		if err != nil {
			return data, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Env.KVToken()))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return data, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return data, err
		}
		return data, nil
	})
}

// Pipeline is a generic command to the KV store.
func Pipeline[Data any](args ...KVCommand) async.Task[[]KVPacket[Data]] {
	return async.New(func() ([]KVPacket[Data], error) {
		commands := make([][]any, len(args))
		for i, arg := range args {
			commands[i] = make([]any, len(arg.Args)+1)
			commands[i][0] = arg.Name
			for j, opt := range arg.Args {
				commands[i][j+1] = opt
			}
		}

		data := []KVPacket[Data]{}
		body, err := json.Marshal(commands)

		if err != nil {
			return data, err
		}

		req, err := http.NewRequest("POST", config.Env.KVURL()+"/pipeline", bytes.NewBuffer(body))
		if err != nil {
			return data, err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Env.KVToken()))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return data, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return data, err
		}
		return data, nil
	})
}

// Get gets a value by their key
func Get(key string) async.Task[KVPacket[string]] {
	return Command[string](get, key)
}

// GetAll gets multiple values by their keys
func GetAll(keys ...string) async.Task[[]KVPacket[string]] {
	args := f.Map(keys, func(key string) KVCommand { return KVCommand{Name: get, Args: []any{key}} })
	return Pipeline[string](args...)
}

// MGet gets multiple values by their keys
func MGet(keys ...string) async.Task[KVPacket[[]string]] {
	args := f.Map(keys, func(key string) any { return key })
	return Command[[]string](mget, args...)
}

// Set sets a value by their key and returns the value
func Set[Data any](key string, value Data) async.Task[KVPacket[Data]] {
	return Command[Data](set, key, value)
}

// Incr increments an integer value by their key and returns the value
// If the key does not exist, it will be created with the value 0 before
func Incr(key string) async.Task[KVPacket[int]] {
	return async.New(func() (KVPacket[int], error) {
		str, err := Command[string](incr, key).Await()
		if err != nil {
			return KVPacket[int]{}, err
		}
		return KVPacket[int]{Result: f.ParseInt(str.Result)}, nil
	})
}
