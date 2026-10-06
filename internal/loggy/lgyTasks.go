package loggy

import (
	"fmt"
	"time"
)

type TaskNotif struct {
	Start    time.Time
	TaskName string
	JobId    int
}

func TaskNotifBegin(taskName string, jobId int, keyvals ...any) TaskNotif {
	Info(fmt.Sprintf(" |-strt-> [%v] |", taskName), append([]any{"JobId", jobId}, keyvals...)...)
	return TaskNotif{
		Start:    time.Now(),
		TaskName: taskName,
		JobId:    jobId,
	}
}

func (notif TaskNotif) End(keyvals ...any) {
	d := time.Since(notif.Start)
	dur := fmt.Sprintf(
		"job completed in %dm %ds",
		int(d.Minutes()),
		int(d.Seconds())%60,
	)
	Info(fmt.Sprintf(" --end-| [%v] |", notif.TaskName), append([]any{"JobId", notif.JobId, "dur", dur}, keyvals...)...)
}
