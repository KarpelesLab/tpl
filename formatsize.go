package tpl

import "fmt"

func FormatSize(v uint64) string {
	var siz = [...]struct {
		unit string
		size float32
	}{
		{unit: "B", size: 1},
		{unit: "kiB", size: 1024},
		{unit: "MiB", size: 1024},
		{unit: "GiB", size: 1024},
		{unit: "TiB", size: 1024},
		{unit: "PiB", size: 1024},
	}

	if v == 0 {
		return "0 B"
	}

	vf := float32(v)
	last := siz[0]
	for _, i := range siz {
		if vf < i.size*1.5 {
			break
		}
		vf = vf / i.size
		last = i
	}

	if last.size > 1 {
		return fmt.Sprintf("%01.2f %s", vf, last.unit)
	} else {
		return fmt.Sprintf("%.0f %s", vf, last.unit)
	}
}
