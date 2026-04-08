package duel

const SelfDamageOnWrong = 5

func ProcessAnswer(attacker, defender *Player, correct bool, speed int) int {
	if correct {
		attacker.CorrectCount++
	} else {
		attacker.WrongCount++
	}

	damage := CalculateDamage(correct, speed)

	if !correct && damage == 0 {
		attacker.HP -= SelfDamageOnWrong
		if attacker.HP < 0 {
			attacker.HP = 0
		}
		return 0
	}

	defender.HP -= damage

	if defender.HP < 0 {
		defender.HP = 0
	}

	return damage
}
