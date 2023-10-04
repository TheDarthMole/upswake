package main

import "upsWake/wol"

func main() {
	err := wol.Wake("00:00:00:00:00:00")
	if err != nil {
		return
	}
}
