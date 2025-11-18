package domain

import "time"

type Route struct {
	ID                    int       `db:"id" json:"id"`
	OriginStationID       int       `db:"origin_station_id" json:"origin_station_id"`
	DestinationStationID  int       `db:"destination_station_id" json:"destination_station_id"`
	RouteHash             string    `db:"route_hash" json:"route_hash"`
	Operators             []string  `db:"operators" json:"operators"`
	TypicalDurationMinutes *int     `db:"typical_duration_minutes" json:"typical_duration_minutes,omitempty"`
	DistanceMiles         *float64  `db:"distance_miles" json:"distance_miles,omitempty"`
	IsTfLRoute            bool      `db:"is_tfl_route" json:"is_tfl_route"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
}

type ServiceRecord struct {
	ID                         int64      `db:"id" json:"id"`
	RouteID                    int        `db:"route_id" json:"route_id"`
	RID                        *string    `db:"rid" json:"rid,omitempty"`
	ServiceDate                time.Time  `db:"service_date" json:"service_date"`
	DayOfWeek                  int        `db:"day_of_week" json:"day_of_week"`
	IsWeekday                  bool       `db:"is_weekday" json:"is_weekday"`
	ScheduledDeparture         string     `db:"scheduled_departure" json:"scheduled_departure"`
	ScheduledArrival           string     `db:"scheduled_arrival" json:"scheduled_arrival"`
	ScheduledDurationMinutes   *int       `db:"scheduled_duration_minutes" json:"scheduled_duration_minutes,omitempty"`
	ActualDeparture            *string    `db:"actual_departure" json:"actual_departure,omitempty"`
	ActualArrival              *string    `db:"actual_arrival" json:"actual_arrival,omitempty"`
	ActualDurationMinutes      *int       `db:"actual_duration_minutes" json:"actual_duration_minutes,omitempty"`
	DepartureDelayMinutes      int        `db:"departure_delay_minutes" json:"departure_delay_minutes"`
	ArrivalDelayMinutes        int        `db:"arrival_delay_minutes" json:"arrival_delay_minutes"`
	WasCancelled               bool       `db:"was_cancelled" json:"was_cancelled"`
	CancellationReason         *string    `db:"cancellation_reason" json:"cancellation_reason,omitempty"`
	DelayReason                *string    `db:"delay_reason" json:"delay_reason,omitempty"`
	IsOnTime                   *bool      `db:"is_on_time" json:"is_on_time,omitempty"`
	IsSlightlyDelayed          *bool      `db:"is_slightly_delayed" json:"is_slightly_delayed,omitempty"`
	IsSignificantlyDelayed     *bool      `db:"is_significantly_delayed" json:"is_significantly_delayed,omitempty"`
	IsSeverelyDelayed          *bool      `db:"is_severely_delayed" json:"is_severely_delayed,omitempty"`
	OperatorCode               *string    `db:"operator_code" json:"operator_code,omitempty"`
	DataSource                 string     `db:"data_source" json:"data_source"`
	CreatedAt                  time.Time  `db:"created_at" json:"created_at"`
}
