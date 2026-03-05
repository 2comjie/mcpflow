import { useState, useEffect, useRef } from 'react'
import { Card, Select, Input, Button, Slider, message, Spin, Tag, Typography } from 'antd'
import { SendOutlined, RobotOutlined, UserOutlined, SettingOutlined } from '@ant-design/icons'
import { llmProviderApi, type LLMProvider } from '../../api/llm_provider'
import { mcpServerApi, type MCPServer } from '../../api/mcpserver'
import { agentApi } from '../../api/agent'
import AgentStepsView from '../../components/AgentStepsView'

const { TextArea } = Input
const { Text } = Typography

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  agent_steps?: any[]
  tool_calls_count?: number
  iterations?: number
  total_tokens?: number
}

export default function AgentPlayground() {
  const [providers, setProviders] = useState<LLMProvider[]>([])
  const [mcpServers, setMcpServers] = useState<MCPServer[]>([])
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null)
  const [selectedServers, setSelectedServers] = useState<string[]>([])
  const [systemMsg, setSystemMsg] = useState('')
  const [temperature, setTemperature] = useState(0.7)
  const [maxTokens, setMaxTokens] = useState(2048)
  const [maxIterations, setMaxIterations] = useState(10)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputValue, setInputValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [showSettings, setShowSettings] = useState(true)
  const chatEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    llmProviderApi.list().then((res: any) => setProviders(res.data || res || []))
    mcpServerApi.list().then((res: any) => setMcpServers(res.data || res || []))
  }, [])

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    if (!inputValue.trim()) return
    if (!selectedProvider) {
      message.warning('Please select an LLM Provider')
      return
    }
    if (selectedServers.length === 0) {
      message.warning('Please select at least one MCP Server')
      return
    }

    const userMsg = inputValue.trim()
    setInputValue('')
    setMessages((prev) => [...prev, { role: 'user', content: userMsg }])
    setLoading(true)

    try {
      const res: any = await agentApi.chat({
        llm_provider_id: selectedProvider,
        mcp_server_ids: selectedServers,
        message: userMsg,
        system_msg: systemMsg || undefined,
        temperature,
        max_tokens: maxTokens,
        max_iterations: maxIterations,
      })
      const data = res.data || res
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: data.content || '',
          agent_steps: data.agent_steps,
          tool_calls_count: data.tool_calls_count,
          iterations: data.iterations,
          total_tokens: data.total_tokens,
        },
      ])
    } catch (err: any) {
      message.error(err.message || 'Agent chat failed')
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: `Error: ${err.message}` },
      ])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ display: 'flex', height: 'calc(100vh - 48px)', gap: 16 }}>
      {/* Left: Settings Panel */}
      <div
        style={{
          width: showSettings ? 320 : 0,
          overflow: 'hidden',
          transition: 'width 0.3s',
          flexShrink: 0,
        }}
      >
        <Card
          title={
            <span>
              <SettingOutlined style={{ marginRight: 8 }} />
              Agent Settings
            </span>
          }
          size="small"
          style={{ height: '100%', overflow: 'auto' }}
        >
          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              LLM Provider *
            </Text>
            <Select
              placeholder="Select LLM Provider"
              style={{ width: '100%' }}
              value={selectedProvider}
              onChange={setSelectedProvider}
              options={providers.map((p) => ({ label: p.name, value: p.id }))}
            />
          </div>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              MCP Servers *
            </Text>
            <Select
              mode="multiple"
              placeholder="Select MCP Servers"
              style={{ width: '100%' }}
              value={selectedServers}
              onChange={setSelectedServers}
              options={mcpServers.map((s) => ({
                label: (
                  <span>
                    {s.name}
                    {s.status === 'active' && (
                      <Tag color="green" style={{ marginLeft: 4, fontSize: 10 }}>
                        active
                      </Tag>
                    )}
                  </span>
                ),
                value: s.id,
              }))}
            />
          </div>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              System Prompt
            </Text>
            <TextArea
              rows={3}
              placeholder="Optional system prompt for the agent..."
              value={systemMsg}
              onChange={(e) => setSystemMsg(e.target.value)}
            />
          </div>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              Temperature: {temperature}
            </Text>
            <Slider min={0} max={2} step={0.1} value={temperature} onChange={setTemperature} />
          </div>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              Max Tokens: {maxTokens}
            </Text>
            <Slider min={256} max={8192} step={256} value={maxTokens} onChange={setMaxTokens} />
          </div>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
              Max Iterations: {maxIterations}
            </Text>
            <Slider min={1} max={20} step={1} value={maxIterations} onChange={setMaxIterations} />
          </div>
        </Card>
      </div>

      {/* Right: Chat Area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Header */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '8px 0',
            marginBottom: 8,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <RobotOutlined style={{ fontSize: 20, color: '#3b5bdb' }} />
            <span style={{ fontSize: 18, fontWeight: 600 }}>Agent Playground</span>
          </div>
          <Button
            type="text"
            icon={<SettingOutlined />}
            onClick={() => setShowSettings(!showSettings)}
          >
            {showSettings ? 'Hide' : 'Show'} Settings
          </Button>
        </div>

        {/* Messages */}
        <div
          style={{
            flex: 1,
            overflow: 'auto',
            padding: '16px 0',
            display: 'flex',
            flexDirection: 'column',
            gap: 16,
          }}
        >
          {messages.length === 0 && (
            <div
              style={{
                flex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexDirection: 'column',
                color: '#98a2b3',
              }}
            >
              <RobotOutlined style={{ fontSize: 48, marginBottom: 16 }} />
              <div style={{ fontSize: 16 }}>Start a conversation with the Agent</div>
              <div style={{ fontSize: 12, marginTop: 4 }}>
                Select LLM Provider and MCP Servers, then send a message
              </div>
            </div>
          )}
          {messages.map((msg, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                gap: 12,
                flexDirection: msg.role === 'user' ? 'row-reverse' : 'row',
              }}
            >
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: '50%',
                  background: msg.role === 'user' ? '#3b5bdb' : '#f0f1f3',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: msg.role === 'user' ? '#fff' : '#667085',
                  flexShrink: 0,
                }}
              >
                {msg.role === 'user' ? <UserOutlined /> : <RobotOutlined />}
              </div>
              <div
                style={{
                  maxWidth: '75%',
                  background: msg.role === 'user' ? '#3b5bdb' : '#fff',
                  color: msg.role === 'user' ? '#fff' : '#1a1a2e',
                  padding: '10px 14px',
                  borderRadius: 12,
                  border: msg.role === 'user' ? 'none' : '1px solid #eaecf0',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  fontSize: 13,
                  lineHeight: 1.6,
                }}
              >
                {msg.content}
                {msg.role === 'assistant' && (
                  <>
                    {(msg.tool_calls_count !== undefined || msg.total_tokens !== undefined) && (
                      <div
                        style={{
                          marginTop: 8,
                          paddingTop: 8,
                          borderTop: '1px solid #eaecf0',
                          display: 'flex',
                          gap: 8,
                          flexWrap: 'wrap',
                        }}
                      >
                        {msg.tool_calls_count !== undefined && msg.tool_calls_count > 0 && (
                          <Tag color="blue" style={{ fontSize: 10, borderRadius: 4 }}>
                            {msg.tool_calls_count} tool calls
                          </Tag>
                        )}
                        {msg.iterations !== undefined && (
                          <Tag style={{ fontSize: 10, borderRadius: 4 }}>
                            {msg.iterations} iterations
                          </Tag>
                        )}
                        {msg.total_tokens !== undefined && (
                          <Tag style={{ fontSize: 10, borderRadius: 4 }}>
                            {msg.total_tokens} tokens
                          </Tag>
                        )}
                      </div>
                    )}
                    {msg.agent_steps && msg.agent_steps.length > 0 && (
                      <div
                        style={{
                          marginTop: 8,
                          paddingTop: 8,
                          borderTop: '1px solid #eaecf0',
                        }}
                      >
                        <AgentStepsView steps={msg.agent_steps} />
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          ))}
          {loading && (
            <div style={{ display: 'flex', gap: 12 }}>
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: '50%',
                  background: '#f0f1f3',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#667085',
                  flexShrink: 0,
                }}
              >
                <RobotOutlined />
              </div>
              <div
                style={{
                  background: '#fff',
                  padding: '10px 14px',
                  borderRadius: 12,
                  border: '1px solid #eaecf0',
                }}
              >
                <Spin size="small" />
                <span style={{ marginLeft: 8, color: '#98a2b3', fontSize: 12 }}>
                  Agent is thinking...
                </span>
              </div>
            </div>
          )}
          <div ref={chatEndRef} />
        </div>

        {/* Input */}
        <div
          style={{
            display: 'flex',
            gap: 8,
            padding: '12px 0',
            borderTop: '1px solid #eaecf0',
          }}
        >
          <Input
            size="large"
            placeholder="Send a message to the agent..."
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onPressEnter={handleSend}
            disabled={loading}
          />
          <Button
            type="primary"
            size="large"
            icon={<SendOutlined />}
            onClick={handleSend}
            loading={loading}
            disabled={!inputValue.trim()}
          >
            Send
          </Button>
        </div>
      </div>
    </div>
  )
}
