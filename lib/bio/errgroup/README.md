This is errgroup from `https://cs.opensource.google/go/x/sync`. It is integrated into the source tree because it doesn't change a lot and is one of the only external dependencies we use, especially for filtering of `bio` parser results.

I modified the code to pass the Go linter requirements.