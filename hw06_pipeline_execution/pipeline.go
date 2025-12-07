package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return in
	}
	current := wrapWithDone(in, done)
	for _, stage := range stages {
		input := current
		current = wrapWithDone(stage(input), done)
	}

	return current
}

func wrapWithDone(in In, done In) Out {
	out := make(Bi)
	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				go drain(in)
				return
			case val, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					go drain(in)
					return
				case out <- val:
				}
			}
		}
	}()

	return out
}

func drain(ch In) {
	for {
		_, ok := <-ch
		if !ok {
			return
		}
	}
}
