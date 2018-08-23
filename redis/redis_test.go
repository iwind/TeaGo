package redis

import "testing"

func TestRedisClient(t *testing.T) {
	client, err := Client()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(client.GetString("hello"))
}
