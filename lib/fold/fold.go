/*
	Package fold contains DNA and RNA folding algorithms.

We have two supported algorithms: The dynamic Zuker algorithm, and LinearFold.

These packages work, but LinearFold needs a massive cleanup effort.
*/
package fold

type SequenceFolder interface {
	Fold(sequence string, temp float64) (dotBracket string, score float64, err error)
}
