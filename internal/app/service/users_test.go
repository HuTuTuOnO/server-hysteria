package service

import (
	"testing"
	"time"
)

func TestTrafficManager_N1(t *testing.T) {
	trafficManager := newTrafficManager()
	trafficItem := newTrafficItem()
	trafficItem.Down.Add(20)
	trafficItem.Up.Add(20)
	trafficItem.Count.Add(20)
	trafficManager.set(1, trafficItem)
	userTraffics := trafficManager.toUserTraffics()

	if len(userTraffics) != 1 {
		t.Error("error")
	}

	trafficManager.clear()
	loadTrafficItem := trafficManager.load(1)
	if loadTrafficItem != nil {
		t.Error("load error")
	}
}

func TestTrafficManager_N2(t *testing.T) {
	trafficManager := newTrafficManager()
	trafficItem := newTrafficItem()
	trafficItem.Down.Add(20)
	trafficItem.Up.Add(20)
	trafficItem.Count.Add(20)
	trafficManager.set(1, trafficItem)

	loadTrafficItem := trafficManager.load(1)
	if loadTrafficItem == nil {
		t.Error("load error")
	}
	loadTrafficItem.Down.Add(20)
	loadTrafficItem.Up.Add(20)
	loadTrafficItem.Count.Add(20)

	userTraffics := trafficManager.toUserTraffics()

	t.Log(userTraffics[0].Upload, userTraffics[0].Download, userTraffics[0].Count)

	if userTraffics[0].Upload != 40 {
		t.Error("up value error")
	}

	if userTraffics[0].Download != 40 {
		t.Error("download value error")
	}

	if userTraffics[0].Count != 40 {
		t.Error("Count value error")
	}
}

func TestTrafficManager_N3(t *testing.T) {
	trafficManager := newTrafficManager()
	trafficItem := newTrafficItem()
	trafficItem.Down.Add(20)
	trafficItem.Up.Add(20)
	trafficItem.Count.Add(20)
	trafficManager.set(1, trafficItem)

	for i := 1; i < 6; i++ {
		go func() {
			loadTrafficItem := trafficManager.load(1)
			if loadTrafficItem == nil {
				t.Error("load error")
			}
			loadTrafficItem.Down.Add(20)
			loadTrafficItem.Up.Add(20)
			loadTrafficItem.Count.Add(20)
		}()
	}

	time.Sleep(1 * time.Second)

	userTraffics := trafficManager.toUserTraffics()
	t.Log(userTraffics[0].Upload, userTraffics[0].Download, userTraffics[0].Count)

	if userTraffics[0].Upload != 120 {
		t.Error("up value error")
	}

	if userTraffics[0].Download != 120 {
		t.Error("download value error")
	}

	if userTraffics[0].Count != 120 {
		t.Error("Count value error")
	}
}
