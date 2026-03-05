import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, Typography } from 'antd';
import {
  ApartmentOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { statsApi } from '../../api/stats';

export default function Dashboard() {
  const [stats, setStats] = useState<any>({});

  useEffect(() => {
    statsApi.get().then((res: any) => setStats(res.data || {}));
  }, []);

  return (
    <div>
      <Typography.Title level={4} style={{ marginBottom: 24 }}>
        系统概览
      </Typography.Title>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card>
            <Statistic
              title="工作流总数"
              value={stats.total_workflows || 0}
              prefix={<ApartmentOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="执行总次数"
              value={stats.total_executions || 0}
              prefix={<PlayCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功次数"
              value={stats.success_executions || 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="失败次数"
              value={stats.failed_executions || 0}
              prefix={<CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="成功率"
              value={stats.success_rate || 0}
              precision={1}
              suffix="%"
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
