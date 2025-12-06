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
	current := in
	for _, stage := range stages {
		input := current
		output := make(Bi)
		go process(stage, done, input, output)
		current = output
	}
	return current
}

func process(s Stage, done In, inChan In, outChan Bi) {
	defer close(outChan)
	stageOut := s(inChan)
	for {
		select {
		case <-done:
			go func() {
				for range stageOut {
				}
			}()
			return
		case val, ok := <-stageOut:
			if !ok {
				return
			}
			select {
			case <-done:
				go func() {
					for range stageOut {
					}
				}()
				return
			case outChan <- val:
			}
		}
	}
}
