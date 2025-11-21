package zuker

/*
These interfaces are here to satisfy the `SequenceFolder` interface, which
looks like this:

type SequenceFolder interface {
    Fold(sequence string, temp float64) (dotBracket string, score float64, error)
}
*/

type ZukerFoldWrapper struct{}

func NewZukerFoldWrapper() ZukerFoldWrapper {
	return ZukerFoldWrapper{}
}

func (z ZukerFoldWrapper) Fold(sequence string, temp float64) (dotBracket string, score float64, err error) {
	result, err := Zuker(sequence, temp)
	if err != nil {
		return "", 0, err
	}
	return result.DotBracket(), result.MinimumFreeEnergy(), nil
}
