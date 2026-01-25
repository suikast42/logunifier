package api

// FeedbackNotifier implements the client.Notifier interface to receive
// callbacks when a batch of logs is successfully sent or fails.
type FeedbackNotifier[T any] struct {
	SendCtx   T
	Entry     *Entry
	OnSuccess func(entry *Entry, sendCtx T, status int)
	OnFailure func(entry *Entry, sendCtx T, status int, err error)
}

// // New creates a new FeedbackNotifier with the provided values.
// // All fields are optional — you can pass nil for any of them.
// func AddFeedbackNotifier[T any](
//
//	entry *Entry,
//	sendCtx T,
//	onSuccess func(*Entry, *T),
//	onFailure func(*Entry, *T),
//
//	) {
//		f := FeedbackNotifier[T]{
//			SendCtx:   &sendCtx,
//			Entry:     entry,
//			OnSuccess: onSuccess,
//			OnFailure: onFailure,
//		}
//		entry.FeedbackNotifier = f
//	}
func AddFeedbackNotifier[T any](
	entry *Entry,
	sendCtx T,
	onSuccess func(*Entry, T, int),
	onFailure func(*Entry, T, int, error),
) {
	f := FeedbackNotifier[any]{
		SendCtx: sendCtx, // compiler error here
		Entry:   entry,
		OnSuccess: func(e *Entry, a any, status int) { // compiler error here
			if onSuccess != nil {
				if v, ok := a.(T); ok {
					onSuccess(e, v, status)
				}
			}
		},
		OnFailure: func(e *Entry, a any, status int, err error) { // compiler error here
			if onFailure != nil {
				if v, ok := a.(T); ok {
					onFailure(e, v, status, err)
				}
			}
		},
	}
	entry.FeedbackNotifier = f
}
