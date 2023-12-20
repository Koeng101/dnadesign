package fold

type SequenceFolder interface {
	Fold(sequence string, temp float64) (dotBracket string, score float64, err error)
}
