package services

import (
	"errors"
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"time"
)

type TimeTrackingService interface {
	ClockIn(userID uint, notes string) (*domain.TimeClock, error)
	ClockOut(userID uint, notes string) (*domain.TimeClock, error)
	StartBreak(userID uint) (*domain.TimeClock, error)
	EndBreak(userID uint) (*domain.TimeClock, error)
	GetLastTimeClock(userID uint) (*domain.TimeClock, error)
	GetLastTimeClockByUsername(username string) (*domain.TimeClock, error)
	GetHistory(userID uint, limit int) ([]domain.TimeClock, error)
	GetTeamStatus() ([]map[string]interface{}, error)
	GetRecentActivities(limit int) ([]domain.TimeClock, error)
	GetWeeklySummary(userID uint) (map[string]interface{}, error)
	GetTeamOverview() (map[string]interface{}, error)
}

type timeTrackingService struct {
	repo     repository.TimeTrackingRepository
	userRepo *repository.UserRepository
}

func NewTimeTrackingService(repo repository.TimeTrackingRepository, userRepo *repository.UserRepository) TimeTrackingService {
	return &timeTrackingService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *timeTrackingService) ClockIn(userID uint, notes string) (*domain.TimeClock, error) {
	last, err := s.repo.GetLastTimeClock(userID)
	if err == nil && last != nil {
		if last.Status == "CLOCKED_IN" || last.Status == "ON_BREAK" {
			return nil, errors.New("already clocked in")
		}
	}

	newEntry := domain.TimeClock{
		UserID:  userID,
		ClockIn: time.Now(),
		Status:  "CLOCKED_IN",
		Notes:   notes,
	}
	if err := s.repo.CreateTimeClock(&newEntry); err != nil {
		return nil, err
	}
	return &newEntry, nil
}

func (s *timeTrackingService) ClockOut(userID uint, notes string) (*domain.TimeClock, error) {
	last, err := s.repo.GetLastTimeClock(userID)
	if err != nil || last == nil || last.Status == "CLOCKED_OUT" {
		return nil, errors.New("not clocked in")
	}

	now := time.Now()
	last.ClockOut = &now
	last.Status = "CLOCKED_OUT"
	if notes != "" {
		last.Notes = last.Notes + " | " + notes
	}

	if last.BreakStart != nil && last.BreakEnd == nil {
		last.BreakEnd = &now
	}

	if err := s.repo.UpdateTimeClock(last); err != nil {
		return nil, err
	}
	return last, nil
}

func (s *timeTrackingService) StartBreak(userID uint) (*domain.TimeClock, error) {
	last, err := s.repo.GetLastTimeClock(userID)
	if err != nil || last == nil {
		return nil, errors.New("not clocked in")
	}
	if last.Status == "CLOCKED_OUT" {
		return nil, errors.New("cannot start break while clocked out")
	}
	if last.Status == "ON_BREAK" {
		return nil, errors.New("already on break")
	}

	now := time.Now()
	last.BreakStart = &now
	last.Status = "ON_BREAK"
	if err := s.repo.UpdateTimeClock(last); err != nil {
		return nil, err
	}
	return last, nil
}

func (s *timeTrackingService) EndBreak(userID uint) (*domain.TimeClock, error) {
	last, err := s.repo.GetLastTimeClock(userID)
	if err != nil || last == nil {
		return nil, errors.New("not clocked in")
	}
	if last.Status != "ON_BREAK" {
		return nil, errors.New("not currently on break")
	}

	now := time.Now()
	last.BreakEnd = &now
	last.Status = "CLOCKED_IN"
	if err := s.repo.UpdateTimeClock(last); err != nil {
		return nil, err
	}
	return last, nil
}

func (s *timeTrackingService) GetLastTimeClock(userID uint) (*domain.TimeClock, error) {
	return s.repo.GetLastTimeClock(userID)
}

func (s *timeTrackingService) GetLastTimeClockByUsername(username string) (*domain.TimeClock, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return s.repo.GetLastTimeClock(user.ID)
}

func (s *timeTrackingService) GetHistory(userID uint, limit int) ([]domain.TimeClock, error) {
	return s.repo.GetRecentTimeClocks(userID, limit)
}

func (s *timeTrackingService) GetTeamStatus() ([]map[string]interface{}, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}

	activeClocks, err := s.repo.GetActiveTimeClocks()
	if err != nil {
		return nil, err
	}

	activeMap := make(map[uint]domain.TimeClock)
	for _, c := range activeClocks {
		activeMap[c.UserID] = c
	}

	var result []map[string]interface{}
	for _, u := range users {
		status := "offline"
		var clockIn time.Time
		var hasClock bool

		if c, ok := activeMap[u.ID]; ok {
			hasClock = true
			clockIn = c.ClockIn
			if c.Status == "CLOCKED_IN" {
				status = "active"
			} else if c.Status == "ON_BREAK" {
				status = "break"
			}
		}

		initials := ""
		if len(u.FirstName) > 0 {
			initials += string(u.FirstName[0])
		}
		if len(u.LastName) > 0 {
			initials += string(u.LastName[0])
		}
		if initials == "" {
			initials = "U"
		}

		result = append(result, map[string]interface{}{
			"id":               u.ID,
			"name":             u.FirstName + " " + u.LastName,
			"role":             u.Role,
			"status":           status,
			"avatar":           initials,
			"clockIn":          clockIn,
			"hasActiveSession": hasClock,
			"productivity":     85 + (int(u.ID) % 15), // Mock
			"hours":            "0h 0m",               // Placeholder
		})
	}
	return result, nil
}

func (s *timeTrackingService) GetRecentActivities(limit int) ([]domain.TimeClock, error) {
	return s.repo.GetAllRecentTimeClocks(limit)
}

func (s *timeTrackingService) GetWeeklySummary(userID uint) (map[string]interface{}, error) {
	start := startOfWeek()
	end := time.Now() // distinct from end of week, we query up to now

	clocks, err := s.repo.GetTimeClocksByDateRange(userID, start, end)
	if err != nil {
		return nil, err
	}

	var totalDuration time.Duration
	dailyMap := make(map[string]time.Duration)
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for _, d := range days {
		dailyMap[d] = 0
	}

	for _, c := range clocks {
		clockEnd := time.Now()
		if c.ClockOut != nil {
			clockEnd = *c.ClockOut
		}

		// If clockIn is before start (shouldn't be based on query), clamp it (optional)
		// Duration
		duration := clockEnd.Sub(c.ClockIn)
		if duration < 0 {
			duration = 0
		} // sanity

		// Add to daily
		dayName := c.ClockIn.Format("Mon")
		dailyMap[dayName] += duration

		// Add to total
		totalDuration += duration
	}

	// Format daily for frontend
	// Frontend expects: stats.goals (Focus Time) -> Today's total

	todayDay := time.Now().Format("Mon")
	todayTotal := dailyMap[todayDay]

	hours := int(totalDuration.Hours())
	minutes := int(totalDuration.Minutes()) % 60

	todayHours := int(todayTotal.Hours())
	todayMinutes := int(todayTotal.Minutes()) % 60

	return map[string]interface{}{
		"weeklyHours":    padTime(hours, minutes),
		"weeklyProgress": (totalDuration.Hours() / 40.0) * 100, // Assuming 40h target
		"todayHours":     padTime(todayHours, todayMinutes),
		"todayProgress":  (todayTotal.Hours() / 8.0) * 100, // Assuming 8h target
	}, nil
}

func (s *timeTrackingService) GetTeamOverview() (map[string]interface{}, error) {
	start := startOfWeek()
	end := time.Now()

	clocks, err := s.repo.GetAllTimeClocksByDateRange(start, end)
	if err != nil {
		return nil, err
	}

	var totalDuration time.Duration
	dailyAttendance := make(map[string]map[uint]bool) // Day -> UserID -> Present
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	for _, d := range days {
		dailyAttendance[d] = make(map[uint]bool)
	}

	for _, c := range clocks {
		clockEnd := time.Now()
		if c.ClockOut != nil {
			clockEnd = *c.ClockOut
		}
		duration := clockEnd.Sub(c.ClockIn)
		totalDuration += duration

		dayIdx := c.ClockIn.Format("Mon")
		dailyAttendance[dayIdx][c.UserID] = true
	}

	// Construct Attendance Graph Data
	// Frontend expects: [{day: 'Mon', present: 8, absent: 0, late: 0}]
	users, _ := s.userRepo.GetAllUsers()
	totalUsers := len(users)

	var attendanceData []map[string]interface{}
	// Order matters? Mon-Sun.
	// Need to strictly order.
	// Map unsorted, slice sorted.

	// Helper for Mon-Sun order
	// Current week might straddle months, but Format("Mon") handles names.
	// We iterate 0 to 6 from start?

	iter := start
	for i := 0; i < 7; i++ {
		dayName := iter.Format("Mon")
		presentCount := len(dailyAttendance[dayName])
		absentCount := totalUsers - presentCount
		// Future days: show 0 logic?
		if iter.After(time.Now()) {
			presentCount = 0
			absentCount = 0
		}

		attendanceData = append(attendanceData, map[string]interface{}{
			"day":     dayName,
			"present": presentCount,
			"absent":  absentCount,
			"late":    0, // Mock "late" or calc based on 9 AM? I'll leave 0 for now as "Late" not defined
		})
		iter = iter.AddDate(0, 0, 1)
	}

	return map[string]interface{}{
		"totalHours": padTime(int(totalDuration.Hours()), int(totalDuration.Minutes())%60),
		"attendance": attendanceData,
		"progress":   (totalDuration.Hours() / (40.0 * float64(totalUsers))) * 100, // Team target
	}, nil
}

func startOfWeek() time.Time {
	now := time.Now()
	offset := int(now.Weekday())
	if offset == 0 {
		offset = 7
	}
	start := now.AddDate(0, 0, -(offset - 1))
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, now.Location())
}

func padTime(h, m int) string {
	return fmt.Sprintf("%dh %02dm", h, m)
}
