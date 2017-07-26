package command

import (
	"github.com/jbgo/sftbot/plx"
)

type ByVolumeDesc []plx.TickerEntry

func (a ByVolumeDesc) Len() int           { return len(a) }
func (a ByVolumeDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVolumeDesc) Less(i, j int) bool { return a[i].BaseVolume > a[j].BaseVolume }
