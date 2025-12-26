package main

// Additional usage examples for the Action Extractor

/*

// === EXAMPLE 1: Processing Slack/Teams Messages ===

func processTeamMessages() {
	messages := []string{
		"@john can you review PR #123 before EOD?",
		"Reminder: Deploy to staging by Wednesday",
		"@all Please fill out the Q1 feedback form",
		"Bug found in checkout flow - needs urgent fix",
	}

	for _, msg := range messages {
		actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: "Extract actions: " + msg},
		})

		for _, action := range actionList.Actions {
			// Send notification if high priority
			if action.Priority == "high" {
				sendSlackNotification(action.Description)
			}
		}
	}
}

// === EXAMPLE 2: Weekly Planning Assistant ===

func weeklyPlanningAssistant() {
	// User's weekly brain dump
	weeklyNotes := `
	This week I need to:
	- Finish the client presentation (due Thursday)
	- Schedule 1-on-1s with team members
	- Review budget proposal
	- Exercise at least 3 times
	- Read that book chapter
	- Buy birthday gift for Sarah (party on Saturday)
	- Car maintenance appointment
	`

	actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Extract weekly actions: " + weeklyNotes},
	})

	// Group by category
	workActions := filterByCategory(actionList.Actions, "work")
	personalActions := filterByCategory(actionList.Actions, "personal")
	healthActions := filterByCategory(actionList.Actions, "health")

	// Create daily plan
	monday := assignToDay(workActions[:2])
	tuesday := assignToDay(workActions[2:])
	// ... continue for other days
}

// === EXAMPLE 3: Meeting Minutes Parser ===

func parseMeetingMinutes() {
	minutes := `
	MEETING: Q4 Planning Session
	Date: 2024-01-15
	Attendees: John, Sarah, Mike, Lisa

	Decisions:
	- Approved budget increase for marketing campaign
	- Decided to hire 2 new engineers in Q2

	Action Items:
	- John: Create job descriptions by next week
	- Sarah: Research marketing agencies, present options in 2 weeks
	- Mike: Set up interviews for candidates
	- Lisa: Update Q4 roadmap in Jira
	- All: Complete skills assessment survey by Friday
	`

	actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Extract meeting action items: " + minutes},
	})

	// Create follow-up email
	email := generateFollowUpEmail(actionList)
	sendEmail("team@company.com", "Meeting Action Items", email)

	// Add to project management tool
	for _, action := range actionList.Actions {
		jira.CreateTask(action.Description, action.DueDate, extractAssignee(action))
	}
}

// === EXAMPLE 4: Smart Email Inbox Assistant ===

func smartInboxAssistant() {
	emails := fetchEmails("INBOX", "UNREAD")

	for _, email := range emails {
		// Extract actions from email body
		actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: "Extract actions from email: " + email.Body},
		})

		if len(actionList.Actions) == 0 {
			// No actions, just archive
			archiveEmail(email)
			continue
		}

		// Has action items
		if actionList.HasDeadline {
			// Add to calendar
			for _, action := range actionList.Actions {
				if action.DueDate != "" {
					calendar.CreateEvent(action.Description, action.DueDate)
				}
			}
		}

		// Tag email
		email.AddLabel("has-actions")
		email.AddLabel(fmt.Sprintf("actions-%d", len(actionList.Actions)))

		// Create tasks
		for _, action := range actionList.Actions {
			todoist.CreateTask(
				action.Description,
				action.Priority,
				action.DueDate,
				action.Tags,
			)
		}
	}
}

// === EXAMPLE 5: Voice Note Transcription Processing ===

func processVoiceNotes() {
	// Assume we have voice-to-text transcription
	transcription := `
	Hey Siri, remind me tomorrow I need to call the dentist to schedule
	that cleaning appointment. Also, I should probably buy flowers for
	mom's birthday next week. Oh and don't let me forget to submit the
	expense report before the end of the month. And I really need to
	start going to the gym regularly, maybe set a goal for 3 times a week.
	`

	actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Extract actions from voice note: " + transcription},
	})

	// Create reminders with appropriate timing
	for _, action := range actionList.Actions {
		reminder := createSmartReminder(action)
		reminderApp.Schedule(reminder)
	}
}

// === EXAMPLE 6: Project Documentation Scanner ===

func scanProjectDocumentation() {
	// Scan all TODO comments in codebase
	todos := findTODOComments("./src")

	// Scan README and CONTRIBUTING files
	readmeContent := readFile("README.md")
	contributingContent := readFile("CONTRIBUTING.md")

	allText := strings.Join(todos, "\n") + readmeContent + contributingContent

	actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
		{Role: roles.User, Content: "Extract all action items and TODOs: " + allText},
	})

	// Create GitHub issues
	for _, action := range actionList.Actions {
		github.CreateIssue(
			repoName,
			action.Description,
			action.Priority,
			action.Tags,
		)
	}
}

// === EXAMPLE 7: Daily Standup Aggregator ===

func aggregateStandupNotes() {
	// Collect standup notes from all team members
	standupNotes := []string{
		"John: Finished API endpoint, will work on tests today, blocked on design feedback",
		"Sarah: Reviewed 3 PRs, starting on user dashboard, no blockers",
		"Mike: Fixed production bug, need to update documentation, waiting for QA approval",
	}

	allActions := []Action{}

	for _, note := range standupNotes {
		actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: "Extract action items: " + note},
		})
		allActions = append(allActions, actionList.Actions...)
	}

	// Generate team summary
	summary := generateTeamActionSummary(allActions)
	sendToSlack("#team-standup", summary)

	// Identify blockers
	blockers := filterBlockers(allActions)
	if len(blockers) > 0 {
		notifyManagement(blockers)
	}
}

// === EXAMPLE 8: Personal Productivity Dashboard ===

func personalProductivityDashboard() {
	sources := map[string]string{
		"emails":     fetchEmailContent(),
		"calendar":   fetchCalendarEvents(),
		"notes":      readNotes("daily.md"),
		"messages":   fetchSlackMessages(),
	}

	dashboard := &ProductivityDashboard{
		Date: time.Now(),
	}

	for source, content := range sources {
		actionList, _, _ := agent.GenerateStructuredData([]messages.Message{
			{Role: roles.User, Content: "Extract actions: " + content},
		})

		dashboard.AddActions(source, actionList.Actions)
	}

	// Analyze productivity
	dashboard.CalculateMetrics()
	dashboard.GenerateReport()

	// Prioritize actions for today
	todaysPriority := dashboard.GetTopPriorityActions(5)
	displayMorningBriefing(todaysPriority)
}

// === Helper Functions ===

func filterByCategory(actions []Action, category string) []Action {
	var filtered []Action
	for _, action := range actions {
		if strings.ToLower(action.Category) == strings.ToLower(category) {
			filtered = append(filtered, action)
		}
	}
	return filtered
}

func extractAssignee(action Action) string {
	// Simple heuristic: look for @mentions in description
	if strings.Contains(action.Description, "@") {
		parts := strings.Split(action.Description, "@")
		if len(parts) > 1 {
			name := strings.Split(parts[1], " ")[0]
			return name
		}
	}
	return "unassigned"
}

func createSmartReminder(action Action) Reminder {
	// Create reminder with smart timing based on action attributes
	reminder := Reminder{
		Title:    action.Description,
		Priority: action.Priority,
	}

	if action.DueDate != "" {
		// Remind 1 day before for high priority
		if action.Priority == "high" {
			reminder.Time = parseDate(action.DueDate).Add(-24 * time.Hour)
		} else {
			reminder.Time = parseDate(action.DueDate).Add(-3 * time.Hour)
		}
	}

	return reminder
}

*/

// To use these examples:
// 1. Uncomment the relevant example
// 2. Adapt to your specific use case
// 3. Integrate with your tools (Slack, Jira, calendar, etc.)
