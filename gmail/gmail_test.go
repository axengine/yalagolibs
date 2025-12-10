package gmail

import (
	"fmt"
	"testing"
)

func TestSendMail(t *testing.T) {
	m := New("smtp.gmail.com", 465, "noreply@yala.org", "")
	msg := fmt.Sprintf(`Receive a new deposit from %s , value %d.
txid:%s`, "alice", 123, "7b8fa812e7e2de09ea461393f91bea8e2461de3227d755f8598af53cf299250e")
	if err := m.SendMail("", []string{"op1@yala.org", "op2@gmail.com"}, "Custody Notify", msg); err != nil {
		t.Error(err)
	}
}
