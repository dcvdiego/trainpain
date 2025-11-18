import { useState } from 'react';
import { Layout, Typography, Card, Alert, Spin } from 'antd';
import { RouteSearchForm } from '../components/search/RouteSearchForm';
import { ReliabilityDashboard } from '../components/dashboard/ReliabilityDashboard';
import { routeService } from '../services/routeService';
import type { RouteReliabilityQuery, RouteReliabilityResponse } from '../types';

const { Header, Content, Footer } = Layout;
const { Title, Text, Paragraph } = Typography;

export const HomePage = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<RouteReliabilityResponse | null>(null);

  const handleSearch = async (query: RouteReliabilityQuery) => {
    setLoading(true);
    setError(null);
    setData(null);

    try {
      const result = await routeService.getReliability(query);
      setData(result);
    } catch (err: any) {
      setError(
        err.response?.data?.error ||
        'Failed to fetch reliability data. Please check your API credentials and try again.'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ background: '#001529', padding: '0 24px' }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: '100%'
        }}>
          <Title level={3} style={{ color: 'white', margin: 0 }}>
            🚂 TrainPain
          </Title>
          <Text style={{ color: 'rgba(255, 255, 255, 0.65)' }}>
            UK Train Reliability Tracker
          </Text>
        </div>
      </Header>

      <Content style={{ padding: '24px', maxWidth: 1200, margin: '0 auto', width: '100%' }}>
        <div style={{ marginBottom: 32 }}>
          <Title level={2} style={{ marginBottom: 8 }}>
            Is your commute reliable? Find out.
          </Title>
          <Paragraph type="secondary" style={{ fontSize: 16 }}>
            Analyze historical UK train performance data to make informed decisions about where to live.
            Check how often trains are cancelled or delayed on your potential commute route.
          </Paragraph>
        </div>

        <Card style={{ marginBottom: 24 }}>
          <RouteSearchForm onSearch={handleSearch} loading={loading} />
        </Card>

        {loading && (
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <Spin size="large" />
            <div style={{ marginTop: 16 }}>
              <Text>Analyzing historical train data...</Text>
            </div>
          </div>
        )}

        {error && !loading && (
          <Alert
            message="Error"
            description={error}
            type="error"
            showIcon
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 24 }}
          />
        )}

        {data && !loading && <ReliabilityDashboard data={data} />}

        {!data && !loading && !error && (
          <Card>
            <div style={{ textAlign: 'center', padding: '40px 20px' }}>
              <Title level={4}>Get Started</Title>
              <Paragraph type="secondary">
                Select your origin and destination stations above to see detailed reliability metrics
                based on real historical performance data.
              </Paragraph>
              <Paragraph type="secondary" style={{ marginTop: 16 }}>
                💡 <strong>Example routes to try:</strong>
                <br />
                Kingston → Waterloo (South West London commute)
                <br />
                Brighton → London Victoria (South Coast to London)
                <br />
                Reading → London Paddington (Thames Valley commute)
              </Paragraph>
            </div>
          </Card>
        )}
      </Content>

      <Footer style={{ textAlign: 'center', background: '#f0f2f5' }}>
        <Paragraph type="secondary" style={{ margin: 0 }}>
          Contains National Rail data © {new Date().getFullYear()}
          <br />
          <Text type="secondary" style={{ fontSize: 12 }}>
            Data is historical and for informational purposes only. Real-time conditions may vary.
          </Text>
        </Paragraph>
      </Footer>
    </Layout>
  );
};
