package duel

func ProcessAnswer(attacker, defender *Player, correct bool, speed int) int {

	damage := CalculateDamage(correct, speed)

	defender.HP -= damage

	if defender.HP < 0 {
		defender.HP = 0
	}

	return damage
}