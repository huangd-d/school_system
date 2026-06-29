import { Card, Col, Row, Statistic } from 'antd'
import {
  BankOutlined,
  BoxPlotOutlined,
  CalendarOutlined,
  DollarOutlined,
} from '@ant-design/icons'

export default function DashboardPage() {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-6">工作台</h2>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="校区总数"
              value={0}
              prefix={<BankOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="物资品类"
              value={0}
              prefix={<BoxPlotOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="进行中活动"
              value={0}
              prefix={<CalendarOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="本月支出"
              value={0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="元"
            />
          </Card>
        </Col>
      </Row>

      <Card className="mt-6" title="欢迎使用">
        <p className="text-gray-500">
          教培机构物资管理系统——轻量级、单机部署的物资全生命周期管理。
        </p>
        <p className="text-gray-400 text-sm mt-2">
          请从左侧菜单选择功能模块开始操作。
        </p>
      </Card>
    </div>
  )
}
