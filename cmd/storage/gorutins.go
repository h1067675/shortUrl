package storage

import (
	"sync"
)

func (s *Storage) generator(chDone chan struct{}, ids struct {
	UserID   int
	LinksIDS []string
}) chan struct {
	userID int
	linkID string
} {
	chRes := make(chan struct {
		userID int
		linkID string
	})
	go func() {
		defer close(chRes)
		for _, e := range ids.LinksIDS {
			select {
			case <-chDone:
				return
			case chRes <- struct {
				userID int
				linkID string
			}{userID: ids.UserID, linkID: e}:
			}
		}
	}()
	return chRes
}

func (s *Storage) fanOut(chDone chan struct{}, chIn chan struct {
	userID int
	linkID string
}, nWorkers int) []chan struct {
	userID int
	linkID int
} {

	channels := make([]chan struct {
		userID int
		linkID int
	}, nWorkers)

	for i := 0; i < nWorkers; i++ {
		addRes := s.checkDletedURL(chDone, chIn)
		channels[i] = addRes
	}
	return channels
}

func (s *Storage) fanIn(chDone chan struct{}, resultChs ...chan struct {
	userID int
	linkID int
}) chan struct {
	userID int
	linkID int
} {
	finalCh := make(chan struct {
		userID int
		linkID int
	})
	var wg sync.WaitGroup
	for _, ch := range resultChs {
		chClosure := ch
		wg.Add(1)

		go func() {
			defer wg.Done()
			for data := range chClosure {
				select {
				case <-chDone:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func (s *Storage) checkDletedURL(chDone chan struct{}, url chan struct {
	userID int
	linkID string
}) chan struct {
	userID int
	linkID int
} {
	chResult := make(chan struct {
		userID int
		linkID int
	})
	go func() {
		defer close(chResult)
		select {
		case <-chDone:
			return
		case in := <-url:
			linkID := s.getUserURLByShortLink(in.userID, in.linkID)
			chResult <- struct {
				userID int
				linkID int
			}{userID: in.userID, linkID: linkID}
		}
	}()
	return chResult
}
