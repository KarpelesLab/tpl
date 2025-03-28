package tpl

import "fmt"

// FormatSize converts a size in bytes to a human-readable format using IEC units.
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
	idx := 0
	for i := 0; i < len(siz); i++ {
		if i > 0 && vf >= siz[i].size {
			vf = vf / siz[i].size
			last = siz[i]
			idx = i
		} else if i == 0 && vf >= 1024 {
			// Special handling for the first conversion
			vf = vf / 1024
			idx = 1
			last = siz[1]
		} else {
			break
		}
	}

	if idx > 0 {
		return fmt.Sprintf("%.2f %s", vf, last.unit)
	} else {
		return fmt.Sprintf("%.0f %s", vf, last.unit)
	}
}
