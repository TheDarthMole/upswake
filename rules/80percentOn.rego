package upswake

# METADATA
# description: Query UPS status and battery charge to determine if the UPS is on mains power & has 80%+ battery charge.
# entrypoint: true
default wake := false

wake if {
	some i
	some j
	some k

	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value >= 80 # 80% or more charge
	input[i].Variables[k].Name == "ups.status"
	input[i].Variables[k].Value == "OL" # On Line (mains is present)
}
