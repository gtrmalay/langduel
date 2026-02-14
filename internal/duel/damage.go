package duel

func CalculateDamage(correct bool, speedMs int) int {

	if !correct {
		return 0
	}

	if speedMs < 2000 {
		return 25
	}

	if speedMs < 4000 {
		return 15
	}

	return 10
}
