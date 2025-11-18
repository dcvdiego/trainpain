import { Card, Row, Col, Statistic, Progress, Typography, Tag, Space, Divider } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { RouteReliabilityResponse } from '../../types/route';
import { getReliabilityLevel, formatPercentage, formatDelay, formatDayFilter } from '../../utils/reliability';

const { Title, Text } = Typography;

interface ReliabilityDashboardProps {
  data: RouteReliabilityResponse;
}

export const ReliabilityDashboard: React.FC<ReliabilityDashboardProps> = ({ data }) => {
  const { origin, destination, metrics, query } = data;
  const reliabilityInfo = getReliabilityLevel(metrics.reliability_score);

  return (
    <div className="reliability-dashboard">
      <Card style={{ marginBottom: 24 }}>
        <Title level={3} style={{ marginBottom: 8 }}>
          {origin.station_name} → {destination.station_name}
        </Title>
        <Text type="secondary">
          {formatDayFilter(query.day_filter)} {query.time_start && query.time_end && `• ${query.time_start} - ${query.time_end}`} • Last {query.analysis_days} days
        </Text>
      </Card>

      {/* Reliability Score */}
      <Card style={{ marginBottom: 24, textAlign: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>
          🎯 Reliability Score
        </Title>
        <div style={{ marginTop: 16 }}>
          <Progress
            type="dashboard"
            percent={metrics.reliability_score}
            strokeColor={reliabilityInfo.color}
            format={(percent) => (
              <div>
                <div style={{ fontSize: 48, fontWeight: 'bold', color: reliabilityInfo.color }}>
                  {Math.round(percent || 0)}
                </div>
                <div style={{ fontSize: 16, marginTop: 8 }}>
                  {reliabilityInfo.level}
                </div>
              </div>
            )}
            size={200}
          />
        </div>
        <Text type="secondary" style={{ display: 'block', marginTop: 16 }}>
          {reliabilityInfo.description}
        </Text>
      </Card>

      {/* Key Metrics */}
      <Card title="📊 Key Metrics" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          <Col xs={12} md={6}>
            <Statistic
              title="On-Time Performance"
              value={formatPercentage(metrics.on_time_rate)}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              Within 5 minutes
            </Text>
          </Col>
          <Col xs={12} md={6}>
            <Statistic
              title="Cancellation Rate"
              value={formatPercentage(metrics.cancellation_rate)}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: metrics.cancellation_rate > 5 ? '#f5222d' : '#faad14' }}
            />
          </Col>
          <Col xs={12} md={6}>
            <Statistic
              title="Average Delay"
              value={formatDelay(metrics.avg_delay_minutes)}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: metrics.avg_delay_minutes > 10 ? '#ff7a45' : '#1890ff' }}
            />
          </Col>
          <Col xs={12} md={6}>
            <Statistic
              title="Worst Case (95th %ile)"
              value={formatDelay(metrics.p95_delay_minutes)}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#ff7a45' }}
            />
          </Col>
        </Row>

        <Divider />

        <div style={{ marginTop: 16 }}>
          <Text strong>Services Analyzed:</Text> {metrics.total_services_analyzed.toLocaleString()} trains
        </div>
      </Card>

      {/* Delay Breakdown */}
      <Card title="📈 Delay Distribution" style={{ marginBottom: 24 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <Text>On-time (0-5 min)</Text>
              <Text strong style={{ color: '#52c41a' }}>
                {formatPercentage(metrics.pct_0_5_min_late)}
              </Text>
            </div>
            <Progress
              percent={metrics.pct_0_5_min_late}
              strokeColor="#52c41a"
              showInfo={false}
            />
          </div>

          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <Text>Slightly delayed (5-15 min)</Text>
              <Text strong style={{ color: '#faad14' }}>
                {formatPercentage(metrics.pct_5_15_min_late)}
              </Text>
            </div>
            <Progress
              percent={metrics.pct_5_15_min_late}
              strokeColor="#faad14"
              showInfo={false}
            />
          </div>

          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <Text>Significantly delayed (15-30 min)</Text>
              <Text strong style={{ color: '#ff7a45' }}>
                {formatPercentage(metrics.pct_15_30_min_late)}
              </Text>
            </div>
            <Progress
              percent={metrics.pct_15_30_min_late}
              strokeColor="#ff7a45"
              showInfo={false}
            />
          </div>

          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <Text>Severely delayed (30+ min)</Text>
              <Text strong style={{ color: '#f5222d' }}>
                {formatPercentage(metrics.pct_30_plus_min_late)}
              </Text>
            </div>
            <Progress
              percent={metrics.pct_30_plus_min_late}
              strokeColor="#f5222d"
              showInfo={false}
            />
          </div>
        </Space>
      </Card>

      {/* Additional Insights */}
      {(metrics.best_day_of_week || metrics.worst_day_of_week) && (
        <Card title="💡 Insights" style={{ marginBottom: 24 }}>
          <Space direction="vertical">
            {metrics.best_day_of_week && (
              <div>
                <Tag color="success">Best Day</Tag>
                <Text>{metrics.best_day_of_week}</Text>
              </div>
            )}
            {metrics.worst_day_of_week && (
              <div>
                <Tag color="error">Worst Day</Tag>
                <Text>{metrics.worst_day_of_week}</Text>
              </div>
            )}
            {metrics.most_common_delay_reason && (
              <div style={{ marginTop: 8 }}>
                <Text type="secondary">Most common delay reason:</Text>
                <br />
                <Text>{metrics.most_common_delay_reason}</Text>
              </div>
            )}
          </Space>
        </Card>
      )}
    </div>
  );
};
