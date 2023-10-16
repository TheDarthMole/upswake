package authz

default allow = false

allow = true {
	input[i].Name == "cyberpower900"
	input[i].Variables[j].Name == "battery.charge"
	input[i].Variables[j].Value >= 80 # 80% or more charge
	input[i].Variables[k].Name == "ups.status"
	input[i].Variables[k].Value == "OL" # On Line (mains is present)
}