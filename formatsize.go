package tpl

import "fmt"

// FormatSize converts a size in bytes to a human-readable format using IEC units.
func FormatSize(v uint64) string {
	units := []string{"B", "kiB", "MiB", "GiB", "TiB", "PiB"}

	if v == 0 {
		return "0 B"
	}

	vf := float64(v)
	idx := 0
	for vf >= 1024 && idx < len(units)-1 {
		vf /= 1024
		idx++
	}

	if idx == 0 {
		return fmt.Sprintf("%.0f %s", vf, units[idx])
	}
	return fmt.Sprintf("%.2f %s", vf, units[idx])
}
