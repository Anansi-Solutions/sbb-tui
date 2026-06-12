package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/necrom4/sbb-tui/model"
)

// buildDetailLines returns the raw text lines that make up a connection's detail view.
func (m appModel) buildDetailLines(c model.Connection, innerWidth int) []string {
	var lines []string

	lines = append(lines, m.renderDisruptionBanner(c)...)

	// Pre-compute label/value widths so platform columns line up across legs.
	labelCol := 0
	valueCol := 0
	for _, leg := range c.Legs {
		if !leg.IsVehicle() {
			continue
		}
		tracks := []string{leg.Track}
		if leg.Exit != nil {
			tracks = append(tracks, leg.Exit.Track)
		}
		for _, p := range tracks {
			if p != "" {
				label := m.icons.platformLabel(p)
				if lw := len([]rune(label)); lw > labelCol {
					labelCol = lw
				}
				if vw := len([]rune(p)); vw > valueCol {
					valueCol = vw
				}
			}
		}
	}
	platformCol := 0
	if labelCol > 0 {
		platformCol = labelCol + 1 + valueCol
	}

	// The last leg of a connection is a bare arrival node; it is rendered
	// through the previous leg's exit, so find the last drawable leg.
	lastRendered := -1
	for i := range c.Legs {
		if c.Legs[i].IsWalk() || c.Legs[i].IsVehicle() {
			lastRendered = i
		}
	}

	for i, leg := range c.Legs {
		isFirst := i == 0
		isLast := i == lastRendered

		switch {
		case leg.IsWalk():
			lines = append(lines, m.renderWalkLeg(leg)...)
		case leg.IsVehicle():
			lines = append(lines, m.renderJourneyLeg(leg, innerWidth, labelCol, platformCol, isFirst, isLast)...)
		default:
			continue
		}

		// Insert blank rows between legs so neighbouring legs don't visually collide.
		if !isLast {
			nextIsWalk := i+1 < len(c.Legs) && c.Legs[i+1].IsWalk()
			currentIsWalk := leg.IsWalk()
			hasArrDelay := leg.IsVehicle() && leg.Exit != nil && leg.Exit.ArrDelay.Minutes > 0
			if currentIsWalk {
				lines = append(lines, "")
			} else if hasArrDelay {
				if nextIsWalk {
				} else {
					lines = append(lines, "")
				}
			} else {
				if nextIsWalk {
					lines = append(lines, "")
				} else {
					lines = append(lines, "", "")
				}
			}
		}
	}

	return lines
}

// renderDisruptionBanner returns the warning lines shown above a
// disrupted connection's detail view, deduplicated by text.
func (m appModel) renderDisruptionBanner(c model.Connection) []string {
	var lines []string
	seen := make(map[string]bool)
	for _, d := range c.DisruptionList() {
		t := d.Text()
		if t.Summary == "" && t.Consequence == "" {
			continue
		}
		key := t.Summary + "|" + t.Consequence
		if seen[key] {
			continue
		}
		seen[key] = true

		lines = append(lines, m.styles.warningBold.Render(m.icons.warning+" "+t.Summary))
		if t.Consequence != "" {
			lines = append(lines, m.styles.warning.Render(t.Consequence))
		}
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	return lines
}

// maxDetailScroll returns the highest scroll offset that still shows new content; 0 means no scrolling needed.
func (m appModel) maxDetailScroll() int {
	if len(m.connections) == 0 || m.resultIndex >= len(m.connections) {
		return 0
	}
	c := m.connections[m.resultIndex]
	boxWidth := max(m.width-borderSize*2-m.resultBoxWidth(), 0)
	innerWidth := max(boxWidth-borderSize-(detailPaddingH*2), 0)

	lines := m.buildDetailLines(c, innerWidth)
	content := strings.Join(lines, "\n")
	wrapped := m.styles.text.Width(innerWidth).Render(content)
	visLines := strings.Split(wrapped, "\n")

	detailFrame := m.styles.detailedResult.GetVerticalFrameSize()
	boxHeight := max(m.resultsHeight()-detailFrame, 0)

	if len(visLines) <= boxHeight {
		return 0
	}
	return len(visLines) - boxHeight
}

// renderFullConnection renders the scrollable detail box for a single connection.
func (m appModel) renderFullConnection(c model.Connection, width int) string {
	innerWidth := max(width-borderSize-(detailPaddingH*2), 0)
	lines := m.buildDetailLines(c, innerWidth)

	detailFrame := m.styles.detailedResult.GetVerticalFrameSize()
	boxHeight := max(m.resultsHeight()-detailFrame, 0)

	// Wrap the content into terminal-width lines so we can scroll by visible row count.
	content := strings.Join(lines, "\n")
	wrapped := m.styles.text.Width(innerWidth).Render(content)
	visLines := strings.Split(wrapped, "\n")

	if len(visLines) > boxHeight {
		scrollY := min(m.detailScrollY, len(visLines)-boxHeight)
		visLines = visLines[scrollY : scrollY+boxHeight]
	}

	return m.styles.detailedResult.Width(width).Height(boxHeight).Render(strings.Join(visLines, "\n"))
}

// renderJourneyLeg renders a single transit leg (departure → vehicle → destination → arrival).
func (m appModel) renderJourneyLeg(leg model.Leg, width, labelCol, platformCol int, isFirst, isLast bool) []string {
	var lines []string

	const timeCol = 5
	const symbolCol = 5

	depTime := leg.Departure.Local().Format("15:04")
	depDelay := leg.DepDelay.Minutes
	depStation := leg.Name
	depPlatform := leg.Track

	depDot := m.icons.hollowDot
	if isFirst {
		depDot = m.icons.filledDot
	}

	depLine := m.formatStationLine(depTime, depDot, depStation, depPlatform, width, timeCol, symbolCol, labelCol, platformCol, true)
	lines = append(lines, depLine)

	indent := strings.Repeat(" ", timeCol)
	spacingLine := fmt.Sprintf("%s  %s", indent, m.icons.vertLine)

	if depDelay > 0 {
		delayStr := m.styles.warningBold.Render(fmt.Sprintf("%*s'", timeCol, fmt.Sprintf("+%d", depDelay)))
		lines = append(lines, fmt.Sprintf("%s %s", delayStr, m.styles.bold.Render(m.icons.vertLine)))
	} else {
		lines = append(lines, spacingLine)
	}

	vehicleIcon := m.styles.vehicleIcon.Render(" " + m.icons.vehicleFor(leg.Type) + " ")
	vehicleModel := m.styles.lineStyle(leg.FgColor, leg.BgColor).Render(leg.DisplayLine())
	company := m.styles.company.Render(leg.OperatorName())
	vehicleLine := fmt.Sprintf("%s  %s  %s %s %s", indent, m.icons.vertLine, vehicleIcon, vehicleModel, company)
	if occupancy := m.renderOccupancy(leg.Occupancy); occupancy != "" {
		vehicleLine += "  " + occupancy
	}
	lines = append(lines, vehicleLine)

	destLine := fmt.Sprintf("%s  %s   %s", indent, m.icons.vertLine, m.styles.textMuted.Render(m.icons.towards+" "+leg.Terminal))
	lines = append(lines, destLine)

	lines = append(lines, spacingLine)

	if leg.Exit != nil {
		arrTime := leg.Exit.Arrival.Local().Format("15:04")
		arrDelay := leg.Exit.ArrDelay.Minutes
		arrStation := leg.Exit.Name
		arrPlatform := leg.Exit.Track

		arrSymbol := m.icons.vertLine
		if isLast {
			arrSymbol = m.icons.filledDot
		}

		arrLine := m.formatStationLine(arrTime, arrSymbol, arrStation, arrPlatform, width, timeCol, symbolCol, labelCol, platformCol, false)
		lines = append(lines, arrLine)

		if arrDelay > 0 {
			delayStr := m.styles.warningBold.Render(fmt.Sprintf("%*s'", timeCol, fmt.Sprintf("+%d", arrDelay)))
			lines = append(lines, delayStr)
		}
	}

	return lines
}

// googleMapsURL returns a Google Maps walking-directions URL between a leg's stop and its exit.
func googleMapsURL(leg model.Leg) string {
	if leg.Exit == nil ||
		(leg.Lat == 0 && leg.Lon == 0) ||
		(leg.Exit.Lat == 0 && leg.Exit.Lon == 0) {
		return ""
	}
	return fmt.Sprintf("https://www.google.com/maps/dir/?api=1&origin=%f,%f&destination=%f,%f&travelmode=walking",
		leg.Lat, leg.Lon, leg.Exit.Lat, leg.Exit.Lon)
}

// renderLink wraps text in an OSC 8 terminal hyperlink.
func renderLink(text, url string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

// walkMinutes returns the duration of a walk leg in minutes.
func walkMinutes(leg model.Leg) int {
	if leg.RunningTime > 0 {
		return leg.RunningTime / 60
	}
	if leg.Exit != nil && !leg.Departure.IsZero() && !leg.Exit.Arrival.IsZero() {
		return int(leg.Exit.Arrival.Sub(leg.Departure).Minutes())
	}
	return 0
}

// renderWalkLeg renders a walking transfer between two stops as a clickable duration.
func (m appModel) renderWalkLeg(leg model.Leg) []string {
	var lines []string

	walkDuration := fmt.Sprintf("%d", walkMinutes(leg))
	if url := googleMapsURL(leg); url != "" {
		// TODO: add `` icon and set that as clickable url link instead of the time
		walkDuration = renderLink(walkDuration, url)
	}

	walkLine := fmt.Sprintf("%s  %s %s'", strings.Repeat(" ", 5), m.icons.walk, walkDuration)
	lines = append(lines, walkLine)

	return lines
}

// formatStationLine formats one row of "time  symbol  station  …  platform"
// in the detail view, padded to align across all legs.
func (m appModel) formatStationLine(timeStr, symbol, station, platform string, width, timeCol, symbolCol, labelCol, platformCol int, bold bool) string {
	textStyle := m.styles.text
	if bold {
		textStyle = m.styles.bold
	}

	timePart := textStyle.Render(timeStr)
	symbolPart := fmt.Sprintf("  %s  ", textStyle.Render(symbol))

	platformPart := ""
	if platform != "" {
		label := m.icons.platformLabel(platform)
		leadingPad := strings.Repeat(" ", max(labelCol-len([]rune(label)), 0))
		labelPart := leadingPad + m.styles.textMuted.Render(label)
		valuePart := textStyle.Render(platform)
		platformPart = labelPart + " " + valuePart
	}

	fixedWidth := timeCol + symbolCol
	if platformCol > 0 {
		fixedWidth += platformCol
	}
	availableForStation := max(width-fixedWidth-1, 5)

	truncatedStation := truncateString(station, availableForStation)
	stationPart := textStyle.Render(truncatedStation)

	stationLen := len([]rune(truncatedStation))
	padding := max(availableForStation-stationLen, 1)

	if platformPart != "" {
		return fmt.Sprintf("%s%s%s%s%s",
			timePart, symbolPart, stationPart, strings.Repeat(" ", padding), platformPart)
	}
	return fmt.Sprintf("%s%s%s", timePart, symbolPart, stationPart)
}

// truncateString shortens s to at most maxLen runes, ending with "..." when it had to clip.
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if maxLen <= 3 {
		return s[:min(len(s), maxLen)]
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// renderSimpleConnection renders one row of the result list with vehicle badge, timeline and platform.
func (m appModel) renderSimpleConnection(c model.Connection, index int, width int) string {
	firstVehicle := c.FirstVehicleLeg()
	lastVehicle := c.LastVehicleLeg()

	style := m.styles.inactive.Width(width)
	if index == m.resultIndex {
		style = m.styles.active.Width(width)
	}

	if firstVehicle == -1 {
		return m.styles.error.Width(width).Padding(1, 2).Render(userError(errConnectionMalformed))
	}

	lineContentWidth := max(width-style.GetHorizontalFrameSize()-2, 0)

	vehicleIcon := m.styles.vehicleIcon.Render(" " + m.icons.vehicleFor(c.Legs[firstVehicle].Type) + " ")
	vehicleModel := m.styles.lineStyle(c.Legs[firstVehicle].FgColor, c.Legs[firstVehicle].BgColor).Render(c.Legs[firstVehicle].DisplayLine())
	company := m.styles.company.Render(c.Legs[firstVehicle].OperatorName())
	endStop := m.styles.text.Render(c.Legs[firstVehicle].Terminal)
	if len(c.DisruptionList()) > 0 {
		endStop += "  " + m.styles.warningBold.Render(m.icons.warning)
	}

	dep := c.Legs[firstVehicle].Departure.Local().Format("15:04")
	arr := c.Arrival.Local().Format("15:04")
	departure := m.styles.bold.Render(dep)
	arrival := m.styles.bold.Render(arr)

	departureDelay := m.formatDelay(c.Legs[firstVehicle].DepDelay.Minutes)
	arrivalDelay := ""
	if exit := c.Legs[lastVehicle].Exit; exit != nil {
		arrivalDelay = m.formatDelay(exit.ArrDelay.Minutes)
	}

	timelinePrefix := ""
	if c.Legs[0].IsWalk() {
		if minutes := walkMinutes(c.Legs[0]); minutes > 0 {
			timelinePrefix = m.icons.walk + " " + m.styles.text.Render(fmt.Sprintf("%d'", minutes)) + "  "
		}
	}

	depGap := "  "
	if departureDelay != "" {
		depGap = " "
	}
	arrGap := "  "
	if arrivalDelay != "" {
		arrGap = " "
	}

	// Size the stops timeline to fill whatever horizontal space is left after the fixed parts.
	timelineFixedWidth := lipgloss.Width(timelinePrefix) +
		lipgloss.Width(departure) +
		lipgloss.Width(departureDelay) + len(depGap) +
		len(arrGap) +
		lipgloss.Width(arrival) +
		lipgloss.Width(arrivalDelay)
	stopsLineWidth := max(lineContentWidth-timelineFixedWidth, stopsLineMinWidth)
	stopsLineRaw := m.renderStopsLine(c, stopsLineWidth)
	timelineWidth := timelineFixedWidth + lipgloss.Width(stopsLineRaw)
	if overflow := timelineWidth - lineContentWidth; overflow > 0 {
		stopsLineWidth = max(stopsLineWidth-overflow, stopsLineMinWidth)
		stopsLineRaw = m.renderStopsLine(c, stopsLineWidth)
	}
	stopsLine := m.styles.bold.Render(stopsLineRaw)

	platformInfo := ""
	platform := c.Legs[firstVehicle].Track
	if platform != "" {
		label := m.icons.platformLabel(platform)
		platformInfo = m.styles.textMuted.Render(label) + " " + m.styles.text.Render(platform)
	}

	duration := m.styles.text.Render(formatDuration(c.Duration))

	occupancy := m.renderOccupancy(c.Occupancy)
	durationPart := duration
	if occupancy != "" {
		durationPart = occupancy + "   " + duration
	}

	bottomLinePadding := max(lineContentWidth-lipgloss.Width(platformInfo)-lipgloss.Width(durationPart), 1)

	content := fmt.Sprintf(
		"\n  %s %s %s  %s\n\n  %s%s%s%s%s  %s%s\n\n  %s%s%v\n",
		vehicleIcon,
		vehicleModel,
		company,
		endStop,
		timelinePrefix,
		departure,
		departureDelay,
		depGap,
		stopsLine,
		arrival,
		arrivalDelay,
		platformInfo,
		strings.Repeat(" ", bottomLinePadding),
		durationPart,
	)

	return style.Render(content)
}

// renderOccupancyClass renders "1." or "2." followed by three person
// glyphs, the first `level` ones bright and the rest dimmed.
// It returns "" when the level is unknown.
func (m appModel) renderOccupancyClass(label string, level int) string {
	if level <= 0 {
		return ""
	}
	level = min(level, 3)

	onStyle := m.styles.text
	if level == 3 {
		onStyle = m.styles.error
	}

	var sb strings.Builder
	sb.WriteString(m.styles.textMuted.Render(label))
	for i := 1; i <= 3; i++ {
		if i <= level {
			sb.WriteString(onStyle.Render(m.icons.person))
		} else {
			sb.WriteString(m.styles.occupancyOff.Render(m.icons.person))
		}
	}
	return sb.String()
}

// renderOccupancy renders the SBB-style per-class occupancy indicator,
// e.g. "1.●○○  2.●●○", or "" when nothing is known.
func (m appModel) renderOccupancy(o model.Occupancy) string {
	first := m.renderOccupancyClass("1.", o.FirstClass())
	second := m.renderOccupancyClass("2.", o.SecondClass())

	switch {
	case first == "" && second == "":
		return ""
	case first == "":
		return second
	case second == "":
		return first
	}
	return first + "  " + second
}

// formatDuration converts the API duration in seconds to "1 h 15 min" or "15 min".
func formatDuration(seconds int) string {
	minutes := seconds / 60
	if minutes >= 60 {
		return fmt.Sprintf("%d h %02d min", minutes/60, minutes%60)
	}
	return fmt.Sprintf("%d min", minutes)
}

// formatDelay returns the styled "+N'" suffix when delay is positive, or an empty string.
func (m appModel) formatDelay(delay int) string {
	if delay > 0 {
		return m.styles.warningBold.Render(fmt.Sprintf(" +%d'", delay))
	}
	return ""
}

// renderStopsLine draws the dotted "●─○─●" timeline between two stations,
// proportional to each leg's duration when available.
func (m appModel) renderStopsLine(c model.Connection, totalWidth int) string {
	if len(c.Legs) == 0 {
		return m.icons.filledDot + m.icons.horizLine + m.icons.horizLine + m.icons.filledDot
	}

	var legDurations []time.Duration
	vehicleCount := 0
	var totalLegDuration time.Duration
	for _, leg := range c.Legs {
		if !leg.IsVehicle() {
			continue
		}
		vehicleCount++
		if leg.Exit == nil {
			continue
		}
		dep := leg.Departure.Time
		arr := leg.Exit.Arrival.Time
		if !dep.IsZero() && !arr.IsZero() {
			dur := arr.Sub(dep)
			legDurations = append(legDurations, dur)
			totalLegDuration += dur
		}
	}

	// Without per-leg durations, distribute hops evenly.
	if totalLegDuration == 0 || len(legDurations) == 0 {
		transfers := max(vehicleCount-1, 0)
		return m.icons.filledDot + strings.Repeat(m.icons.horizLine+m.icons.horizLine+m.icons.hollowDot, transfers) + m.icons.horizLine + m.icons.horizLine + m.icons.filledDot
	}

	var sb strings.Builder
	sb.WriteString(m.icons.filledDot)

	usedChars := 0
	for i, legDur := range legDurations {
		var lineChars int
		if i == len(legDurations)-1 {
			// Give the last leg whatever rounding remainder is left so the line ends flush.
			lineChars = totalWidth - usedChars
		} else {
			proportion := float64(legDur) / float64(totalLegDuration)
			lineChars = int(proportion*float64(totalWidth) + 0.5)
		}
		lineChars = max(lineChars, 1)
		usedChars += lineChars

		sb.WriteString(strings.Repeat(m.icons.horizLine, lineChars))
		if i < len(legDurations)-1 {
			sb.WriteString(m.icons.hollowDot)
		} else {
			sb.WriteString(m.icons.filledDot)
		}
	}

	return sb.String()
}
