export const getReliabilityLevel = (score: number): { level: string; color: string; description: string } => {
  if (score >= 90) {
    return {
      level: 'Excellent',
      color: '#52c41a', // success green
      description: 'Highly reliable - Expect consistent on-time performance',
    };
  } else if (score >= 75) {
    return {
      level: 'Good',
      color: '#1890ff', // primary blue
      description: 'Generally reliable - Occasional minor delays',
    };
  } else if (score >= 60) {
    return {
      level: 'Fair',
      color: '#faad14', // warning orange
      description: 'Moderate reliability - Plan buffer time',
    };
  } else if (score >= 40) {
    return {
      level: 'Poor',
      color: '#ff7a45', // error light orange
      description: 'Unreliable - Expect frequent delays',
    };
  } else {
    return {
      level: 'Very Poor',
      color: '#f5222d', // error red
      description: 'Very unreliable - Consider alternative routes',
    };
  }
};

export const formatPercentage = (value: number): string => {
  return `${value.toFixed(1)}%`;
};

export const formatDelay = (minutes: number): string => {
  if (minutes < 1) {
    return 'On time';
  }
  return `${Math.round(minutes)} min`;
};

export const formatDayFilter = (filter: string): string => {
  switch (filter) {
    case 'weekday':
      return 'Weekdays';
    case 'weekend':
      return 'Weekends';
    case 'monday':
      return 'Mondays';
    case 'tuesday':
      return 'Tuesdays';
    case 'wednesday':
      return 'Wednesdays';
    case 'thursday':
      return 'Thursdays';
    case 'friday':
      return 'Fridays';
    case 'saturday':
      return 'Saturdays';
    case 'sunday':
      return 'Sundays';
    default:
      return 'All days';
  }
};
