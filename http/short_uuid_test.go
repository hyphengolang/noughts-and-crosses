package http

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
)

func TestShortUUID(t *testing.T) {
	uid := uuid.New()
	sid := shortuuid.DefaultEncoder.Encode(uid)

	tid, err := shortuuid.DefaultEncoder.Decode(sid)
	if err != nil {
		t.Fatal(err)
	}

	if tid.String() != uid.String() {
		t.Fatalf("got %s; expected %s\n", tid.String(), uid.String())
	}

	fmt.Println(tid)
	fmt.Println(uid)
}
