package cgroups_topics_page

import (
	"fmt"
	"ktea/kadmin"
	"ktea/tests"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestZZDebug(t *testing.T) {
	model, _ := New(kadmin.NewMockKadmin(), "test-group")
	var o []kadmin.TopicPartitionOffset
	for i := 0; i < 3; i++ {
		o = append(o, kadmin.TopicPartitionOffset{Topic: "topic-1", Partition: int32(i), Offset: 10, HighWaterMark: 10})
	}
	model.Update(kadmin.OffsetListedMsg{Offsets: o})
	fmt.Println(ansi.Strip(model.View(tests.NewKontext(), tests.Renderer)))
}
