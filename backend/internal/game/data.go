package game

import "math/rand"

func cardsDataset() []Card {
	return []Card{
		{ID: "1_black_01", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 1, Green: 1, Red: 1, White: 1}},
		{ID: "1_black_02", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 2, Green: 1, Red: 1, White: 1}},
		{ID: "1_black_03", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 2, Green: 0, Red: 1, White: 2}},
		{ID: "1_black_04", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 1, Blue: 0, Green: 1, Red: 3, White: 0}},
		{ID: "1_black_05", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 2, Red: 1, White: 0}},
		{ID: "1_black_06", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 2, Red: 0, White: 2}},
		{ID: "1_black_07", Tier: 1, Bonus: GemBlack, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 3, Red: 0, White: 0}},
		{ID: "1_black_08", Tier: 1, Bonus: GemBlack, Points: 1, Cost: TokenSet{Black: 0, Blue: 4, Green: 0, Red: 0, White: 0}},
		{ID: "1_blue_01", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 1, Blue: 0, Green: 1, Red: 1, White: 1}},
		{ID: "1_blue_02", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 1, Blue: 0, Green: 1, Red: 2, White: 1}},
		{ID: "1_blue_03", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 2, Red: 2, White: 1}},
		{ID: "1_blue_04", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 0, Blue: 1, Green: 3, Red: 1, White: 0}},
		{ID: "1_blue_05", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 2, Blue: 0, Green: 0, Red: 0, White: 1}},
		{ID: "1_blue_06", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 2, Blue: 0, Green: 2, Red: 0, White: 0}},
		{ID: "1_blue_07", Tier: 1, Bonus: GemBlue, Points: 0, Cost: TokenSet{Black: 3, Blue: 0, Green: 0, Red: 0, White: 0}},
		{ID: "1_blue_08", Tier: 1, Bonus: GemBlue, Points: 1, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 4, White: 0}},
		{ID: "1_white_01", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 1, Red: 1, White: 0}},
		{ID: "1_white_02", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 2, Red: 1, White: 0}},
		{ID: "1_white_03", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 1, Blue: 2, Green: 2, Red: 0, White: 0}},
		{ID: "1_white_04", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 0, Red: 0, White: 3}},
		{ID: "1_white_05", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 1, Blue: 0, Green: 0, Red: 2, White: 0}},
		{ID: "1_white_06", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 2, Blue: 2, Green: 0, Red: 0, White: 0}},
		{ID: "1_white_07", Tier: 1, Bonus: GemWhite, Points: 0, Cost: TokenSet{Black: 0, Blue: 3, Green: 0, Red: 0, White: 0}},
		{ID: "1_white_08", Tier: 1, Bonus: GemWhite, Points: 1, Cost: TokenSet{Black: 0, Blue: 0, Green: 4, Red: 0, White: 0}},
		{ID: "1_green_01", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 0, Red: 1, White: 1}},
		{ID: "1_green_02", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 2, Blue: 1, Green: 0, Red: 1, White: 1}},
		{ID: "1_green_03", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 2, Blue: 1, Green: 0, Red: 2, White: 0}},
		{ID: "1_green_04", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 0, Blue: 3, Green: 1, Red: 0, White: 1}},
		{ID: "1_green_05", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 0, Blue: 1, Green: 0, Red: 0, White: 2}},
		{ID: "1_green_06", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 0, Blue: 2, Green: 0, Red: 2, White: 0}},
		{ID: "1_green_07", Tier: 1, Bonus: GemGreen, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 3, White: 0}},
		{ID: "1_green_08", Tier: 1, Bonus: GemGreen, Points: 1, Cost: TokenSet{Black: 4, Blue: 0, Green: 0, Red: 0, White: 0}},
		{ID: "1_red_01", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 1, Red: 0, White: 1}},
		{ID: "1_red_02", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 1, Blue: 1, Green: 1, Red: 0, White: 2}},
		{ID: "1_red_03", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 2, Blue: 0, Green: 1, Red: 0, White: 2}},
		{ID: "1_red_04", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 3, Blue: 0, Green: 0, Red: 1, White: 1}},
		{ID: "1_red_05", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 0, Blue: 2, Green: 1, Red: 0, White: 0}},
		{ID: "1_red_06", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 2, White: 2}},
		{ID: "1_red_07", Tier: 1, Bonus: GemRed, Points: 0, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 0, White: 3}},
		{ID: "1_red_08", Tier: 1, Bonus: GemRed, Points: 1, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 0, White: 4}},
		{ID: "2_black_01", Tier: 2, Bonus: GemBlack, Points: 1, Cost: TokenSet{Black: 0, Blue: 2, Green: 2, Red: 0, White: 3}},
		{ID: "2_black_02", Tier: 2, Bonus: GemBlack, Points: 1, Cost: TokenSet{Black: 2, Blue: 0, Green: 3, Red: 0, White: 3}},
		{ID: "2_black_03", Tier: 2, Bonus: GemBlack, Points: 2, Cost: TokenSet{Black: 0, Blue: 1, Green: 4, Red: 2, White: 0}},
		{ID: "2_black_04", Tier: 2, Bonus: GemBlack, Points: 2, Cost: TokenSet{Black: 0, Blue: 0, Green: 5, Red: 3, White: 0}},
		{ID: "2_black_05", Tier: 2, Bonus: GemBlack, Points: 2, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 0, White: 5}},
		{ID: "2_black_06", Tier: 2, Bonus: GemBlack, Points: 3, Cost: TokenSet{Black: 6, Blue: 0, Green: 0, Red: 0, White: 0}},
		{ID: "2_blue_01", Tier: 2, Bonus: GemBlue, Points: 1, Cost: TokenSet{Black: 0, Blue: 2, Green: 2, Red: 3, White: 0}},
		{ID: "2_blue_02", Tier: 2, Bonus: GemBlue, Points: 1, Cost: TokenSet{Black: 3, Blue: 2, Green: 3, Red: 0, White: 0}},
		{ID: "2_blue_03", Tier: 2, Bonus: GemBlue, Points: 2, Cost: TokenSet{Black: 0, Blue: 3, Green: 0, Red: 0, White: 5}},
		{ID: "2_blue_04", Tier: 2, Bonus: GemBlue, Points: 2, Cost: TokenSet{Black: 4, Blue: 0, Green: 0, Red: 1, White: 2}},
		{ID: "2_blue_05", Tier: 2, Bonus: GemBlue, Points: 2, Cost: TokenSet{Black: 0, Blue: 5, Green: 0, Red: 0, White: 0}},
		{ID: "2_blue_06", Tier: 2, Bonus: GemBlue, Points: 3, Cost: TokenSet{Black: 0, Blue: 6, Green: 0, Red: 0, White: 0}},
		{ID: "2_white_01", Tier: 2, Bonus: GemWhite, Points: 1, Cost: TokenSet{Black: 2, Blue: 0, Green: 3, Red: 2, White: 0}},
		{ID: "2_white_02", Tier: 2, Bonus: GemWhite, Points: 1, Cost: TokenSet{Black: 0, Blue: 3, Green: 0, Red: 3, White: 2}},
		{ID: "2_white_03", Tier: 2, Bonus: GemWhite, Points: 2, Cost: TokenSet{Black: 2, Blue: 0, Green: 1, Red: 4, White: 0}},
		{ID: "2_white_04", Tier: 2, Bonus: GemWhite, Points: 2, Cost: TokenSet{Black: 3, Blue: 0, Green: 0, Red: 5, White: 0}},
		{ID: "2_white_05", Tier: 2, Bonus: GemWhite, Points: 2, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 5, White: 0}},
		{ID: "2_white_06", Tier: 2, Bonus: GemWhite, Points: 3, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 0, White: 6}},
		{ID: "2_green_01", Tier: 2, Bonus: GemGreen, Points: 1, Cost: TokenSet{Black: 0, Blue: 0, Green: 2, Red: 3, White: 3}},
		{ID: "2_green_02", Tier: 2, Bonus: GemGreen, Points: 1, Cost: TokenSet{Black: 2, Blue: 3, Green: 0, Red: 0, White: 2}},
		{ID: "2_green_03", Tier: 2, Bonus: GemGreen, Points: 2, Cost: TokenSet{Black: 1, Blue: 2, Green: 0, Red: 0, White: 4}},
		{ID: "2_green_04", Tier: 2, Bonus: GemGreen, Points: 2, Cost: TokenSet{Black: 0, Blue: 5, Green: 3, Red: 0, White: 0}},
		{ID: "2_green_05", Tier: 2, Bonus: GemGreen, Points: 2, Cost: TokenSet{Black: 0, Blue: 0, Green: 5, Red: 0, White: 0}},
		{ID: "2_green_06", Tier: 2, Bonus: GemGreen, Points: 3, Cost: TokenSet{Black: 0, Blue: 0, Green: 6, Red: 0, White: 0}},
		{ID: "2_red_01", Tier: 2, Bonus: GemRed, Points: 1, Cost: TokenSet{Black: 3, Blue: 0, Green: 0, Red: 2, White: 2}},
		{ID: "2_red_02", Tier: 2, Bonus: GemRed, Points: 1, Cost: TokenSet{Black: 3, Blue: 3, Green: 0, Red: 2, White: 0}},
		{ID: "2_red_03", Tier: 2, Bonus: GemRed, Points: 2, Cost: TokenSet{Black: 0, Blue: 4, Green: 2, Red: 0, White: 1}},
		{ID: "2_red_04", Tier: 2, Bonus: GemRed, Points: 2, Cost: TokenSet{Black: 5, Blue: 0, Green: 0, Red: 0, White: 3}},
		{ID: "2_red_05", Tier: 2, Bonus: GemRed, Points: 2, Cost: TokenSet{Black: 5, Blue: 0, Green: 0, Red: 0, White: 0}},
		{ID: "2_red_06", Tier: 2, Bonus: GemRed, Points: 3, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 6, White: 0}},
		{ID: "3_black_01", Tier: 3, Bonus: GemBlack, Points: 3, Cost: TokenSet{Black: 0, Blue: 3, Green: 5, Red: 3, White: 3}},
		{ID: "3_black_02", Tier: 3, Bonus: GemBlack, Points: 4, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 7, White: 0}},
		{ID: "3_black_03", Tier: 3, Bonus: GemBlack, Points: 4, Cost: TokenSet{Black: 3, Blue: 0, Green: 3, Red: 6, White: 0}},
		{ID: "3_black_04", Tier: 3, Bonus: GemBlack, Points: 5, Cost: TokenSet{Black: 3, Blue: 0, Green: 0, Red: 7, White: 0}},
		{ID: "3_blue_01", Tier: 3, Bonus: GemBlue, Points: 3, Cost: TokenSet{Black: 5, Blue: 0, Green: 3, Red: 3, White: 3}},
		{ID: "3_blue_02", Tier: 3, Bonus: GemBlue, Points: 4, Cost: TokenSet{Black: 0, Blue: 0, Green: 0, Red: 0, White: 7}},
		{ID: "3_blue_03", Tier: 3, Bonus: GemBlue, Points: 4, Cost: TokenSet{Black: 3, Blue: 3, Green: 0, Red: 0, White: 6}},
		{ID: "3_blue_04", Tier: 3, Bonus: GemBlue, Points: 5, Cost: TokenSet{Black: 0, Blue: 3, Green: 0, Red: 0, White: 7}},
		{ID: "3_white_01", Tier: 3, Bonus: GemWhite, Points: 3, Cost: TokenSet{Black: 3, Blue: 3, Green: 3, Red: 5, White: 0}},
		{ID: "3_white_02", Tier: 3, Bonus: GemWhite, Points: 4, Cost: TokenSet{Black: 7, Blue: 0, Green: 0, Red: 0, White: 0}},
		{ID: "3_white_03", Tier: 3, Bonus: GemWhite, Points: 4, Cost: TokenSet{Black: 6, Blue: 0, Green: 0, Red: 3, White: 3}},
		{ID: "3_white_04", Tier: 3, Bonus: GemWhite, Points: 5, Cost: TokenSet{Black: 7, Blue: 0, Green: 0, Red: 0, White: 3}},
		{ID: "3_green_01", Tier: 3, Bonus: GemGreen, Points: 3, Cost: TokenSet{Black: 3, Blue: 3, Green: 0, Red: 3, White: 5}},
		{ID: "3_green_02", Tier: 3, Bonus: GemGreen, Points: 4, Cost: TokenSet{Black: 0, Blue: 7, Green: 0, Red: 0, White: 0}},
		{ID: "3_green_03", Tier: 3, Bonus: GemGreen, Points: 4, Cost: TokenSet{Black: 0, Blue: 6, Green: 3, Red: 0, White: 3}},
		{ID: "3_green_04", Tier: 3, Bonus: GemGreen, Points: 5, Cost: TokenSet{Black: 0, Blue: 7, Green: 3, Red: 0, White: 0}},
		{ID: "3_red_01", Tier: 3, Bonus: GemRed, Points: 3, Cost: TokenSet{Black: 3, Blue: 5, Green: 3, Red: 0, White: 3}},
		{ID: "3_red_02", Tier: 3, Bonus: GemRed, Points: 4, Cost: TokenSet{Black: 0, Blue: 0, Green: 7, Red: 0, White: 0}},
		{ID: "3_red_03", Tier: 3, Bonus: GemRed, Points: 4, Cost: TokenSet{Black: 0, Blue: 3, Green: 6, Red: 3, White: 0}},
		{ID: "3_red_04", Tier: 3, Bonus: GemRed, Points: 5, Cost: TokenSet{Black: 0, Blue: 0, Green: 7, Red: 3, White: 0}},
	}
}

func noblesDataset() []Noble {
	return []Noble{
		{ID: "n1", Points: 3, Requirement: TokenSet{White: 4, Blue: 4}},
		{ID: "n2", Points: 3, Requirement: TokenSet{White: 4, Green: 4}},
		{ID: "n3", Points: 3, Requirement: TokenSet{White: 4, Red: 4}},
		{ID: "n4", Points: 3, Requirement: TokenSet{White: 4, Black: 4}},
		{ID: "n5", Points: 3, Requirement: TokenSet{Blue: 4, Green: 4}},
		{ID: "n6", Points: 3, Requirement: TokenSet{Blue: 4, Red: 4}},
		{ID: "n7", Points: 3, Requirement: TokenSet{Blue: 4, Black: 4}},
		{ID: "n8", Points: 3, Requirement: TokenSet{Green: 4, Red: 4}},
		{ID: "n9", Points: 3, Requirement: TokenSet{Green: 4, Black: 4}},
		{ID: "n10", Points: 3, Requirement: TokenSet{Red: 4, Black: 4}},
	}
}

func initDecks() (deck1 []Card, deck2 []Card, deck3 []Card) {
	for _, c := range cardsDataset() {
		switch c.Tier {
		case 1:
			deck1 = append(deck1, c)
		case 2:
			deck2 = append(deck2, c)
		case 3:
			deck3 = append(deck3, c)
		}
	}

	rand.Shuffle(len(deck1), func(i, j int) { deck1[i], deck1[j] = deck1[j], deck1[i] })
	rand.Shuffle(len(deck2), func(i, j int) { deck2[i], deck2[j] = deck2[j], deck2[i] })
	rand.Shuffle(len(deck3), func(i, j int) { deck3[i], deck3[j] = deck3[j], deck3[i] })
	return deck1, deck2, deck3
}

func draw(deck *[]Card, n int) []Card {
	if n <= 0 || len(*deck) == 0 {
		return nil
	}
	if n > len(*deck) {
		n = len(*deck)
	}
	out := append([]Card(nil), (*deck)[:n]...)
	*deck = (*deck)[n:]
	return out
}
