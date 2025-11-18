import { useState } from 'react';
import { Form, Select, Button, Row, Col, message, TimePicker } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { stationService } from '../../services/stationService';
import type { StationSearchResult, RouteReliabilityQuery } from '../../types';
import dayjs from 'dayjs';

const { Option } = Select;

interface RouteSearchFormProps {
  onSearch: (query: RouteReliabilityQuery) => void;
  loading?: boolean;
}

export const RouteSearchForm: React.FC<RouteSearchFormProps> = ({ onSearch, loading }) => {
  const [form] = Form.useForm();
  const [originOptions, setOriginOptions] = useState<StationSearchResult[]>([]);
  const [destinationOptions, setDestinationOptions] = useState<StationSearchResult[]>([]);
  const [searchingOrigin, setSearchingOrigin] = useState(false);
  const [searchingDestination, setSearchingDestination] = useState(false);

  const handleOriginSearch = async (value: string) => {
    if (value.length < 2) {
      setOriginOptions([]);
      return;
    }

    setSearchingOrigin(true);
    try {
      const results = await stationService.search(value);
      setOriginOptions(results);
    } catch (error) {
      message.error('Failed to search stations');
    } finally {
      setSearchingOrigin(false);
    }
  };

  const handleDestinationSearch = async (value: string) => {
    if (value.length < 2) {
      setDestinationOptions([]);
      return;
    }

    setSearchingDestination(true);
    try {
      const results = await stationService.search(value);
      setDestinationOptions(results);
    } catch (error) {
      message.error('Failed to search stations');
    } finally {
      setSearchingDestination(false);
    }
  };

  const handleSubmit = (values: any) => {
    const query: RouteReliabilityQuery = {
      origin_crs: values.origin,
      destination_crs: values.destination,
      time_start: values.time_range?.[0]?.format('HH:mm') || undefined,
      time_end: values.time_range?.[1]?.format('HH:mm') || undefined,
      day_filter: values.day_filter || 'all',
      analysis_days: values.analysis_days || 90,
    };

    onSearch(query);
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{
        day_filter: 'weekday',
        analysis_days: 30,
        time_range: [dayjs('07:00', 'HH:mm'), dayjs('09:00', 'HH:mm')],
      }}
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Form.Item
            label="From"
            name="origin"
            rules={[{ required: true, message: 'Please select origin station' }]}
          >
            <Select
              showSearch
              placeholder="Start typing station name..."
              filterOption={false}
              onSearch={handleOriginSearch}
              loading={searchingOrigin}
              notFoundContent={searchingOrigin ? 'Searching...' : 'Type to search stations'}
              size="large"
            >
              {originOptions.map((station) => (
                <Option key={station.id} value={station.crs_code || ''}>
                  {station.station_name} {station.crs_code && `(${station.crs_code})`}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>

        <Col xs={24} md={12}>
          <Form.Item
            label="To"
            name="destination"
            rules={[{ required: true, message: 'Please select destination station' }]}
          >
            <Select
              showSearch
              placeholder="Start typing station name..."
              filterOption={false}
              onSearch={handleDestinationSearch}
              loading={searchingDestination}
              notFoundContent={searchingDestination ? 'Searching...' : 'Type to search stations'}
              size="large"
            >
              {destinationOptions.map((station) => (
                <Option key={station.id} value={station.crs_code || ''}>
                  {station.station_name} {station.crs_code && `(${station.crs_code})`}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Col>

        <Col xs={24} md={8}>
          <Form.Item label="Day Filter" name="day_filter">
            <Select size="large">
              <Option value="all">All Days</Option>
              <Option value="weekday">Weekdays</Option>
              <Option value="weekend">Weekends</Option>
              <Option value="monday">Mondays</Option>
              <Option value="tuesday">Tuesdays</Option>
              <Option value="wednesday">Wednesdays</Option>
              <Option value="thursday">Thursdays</Option>
              <Option value="friday">Fridays</Option>
              <Option value="saturday">Saturdays</Option>
              <Option value="sunday">Sundays</Option>
            </Select>
          </Form.Item>
        </Col>

        <Col xs={24} md={6}>
          <Form.Item label="Time Window" name="time_range">
            <TimePicker.RangePicker
              format="HH:mm"
              size="large"
              minuteStep={15}
            />
          </Form.Item>
        </Col>

        <Col xs={24} md={6}>
          <Form.Item label="Analysis Period" name="analysis_days">
            <Select size="large">
              <Option value={7}>Last 7 days</Option>
              <Option value={30}>Last 30 days</Option>
              <Option value={60}>Last 60 days</Option>
              <Option value={90}>Last 90 days</Option>
            </Select>
          </Form.Item>
        </Col>

        <Col xs={24} md={4}>
          <Form.Item label=" ">
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              icon={<SearchOutlined />}
              loading={loading}
              block
            >
              Search
            </Button>
          </Form.Item>
        </Col>
      </Row>
    </Form>
  );
};
