#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MCPFlow -- 3-Part PPT
Part1: MCP + Workflow Background
Part2: Design (Architecture, Node Schema, Engine, MCP Integration)
Part3: Screenshot Demo (1 per page) + Thanks
"""

from pptx import Presentation
from pptx.util import Inches, Pt
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN
from pptx.enum.shapes import MSO_SHAPE, MSO_CONNECTOR_TYPE
from PIL import Image
import os

# ===== COLORS =====
DARK_BG   = RGBColor(0x0D, 0x1B, 0x2A)
DEEPER    = RGBColor(0x06, 0x0F, 0x1E)
CARD_BG   = RGBColor(0x1B, 0x2E, 0x4A)
ACCENT    = RGBColor(0x00, 0x96, 0xD6)
LIGHT_B   = RGBColor(0x48, 0xCA, 0xE4)
WHITE     = RGBColor(0xFF, 0xFF, 0xFF)
GRAY      = RGBColor(0xC8, 0xD6, 0xE5)
ORANGE    = RGBColor(0xFF, 0xA0, 0x50)
GREEN     = RGBColor(0x4E, 0xC9, 0x8F)
BORDER    = RGBColor(0x2A, 0x4A, 0x6B)
PURPLE    = RGBColor(0xA0, 0x80, 0xFF)
SALMON    = RGBColor(0xF0, 0x80, 0x60)
LIGHT_FILL= RGBColor(0x50, 0x80, 0xF0)
DARK_CARD = RGBColor(0x20, 0x35, 0x55)
YELLOW    = RGBColor(0xFF, 0xD7, 0x00)
RED       = RGBColor(0xFF, 0x60, 0x60)

prs = Presentation()
prs.slide_width  = Inches(13.333)
prs.slide_height = Inches(7.5)
BASE = os.path.dirname(os.path.abspath(__file__))
IMGS = os.path.join(BASE, 'imgs')

# ===== HELPERS =====
def bg(s, c1=DARK_BG, c2=None):
    if c2:
        fill = s.background.fill; fill.gradient(); fill.gradient_angle = 135.0
        fill.gradient_stops[0].color.rgb = c1
        fill.gradient_stops[1].color.rgb = c2
    else:
        s.background.fill.solid(); s.background.fill.fore_color.rgb = c1

def rect(s, l, t, w, h, fill=CARD_BG, bc=None, r=0.05):
    sh = s.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, l, t, w, h)
    sh.fill.solid(); sh.fill.fore_color.rgb = fill
    if bc: sh.line.color.rgb = bc; sh.line.width = Pt(1)
    else: sh.line.fill.background()
    if r > 0 and hasattr(sh, 'adjustments') and len(sh.adjustments) > 0:
        sh.adjustments[0] = r
    return sh

def plain_rect(s, l, t, w, h, fill=CARD_BG, bc=None):
    sh = s.shapes.add_shape(MSO_SHAPE.RECTANGLE, l, t, w, h)
    sh.fill.solid(); sh.fill.fore_color.rgb = fill
    if bc: sh.line.color.rgb = bc; sh.line.width = Pt(1)
    else: sh.line.fill.background()
    return sh

def tb(s, l, t, w, h, text, fs=12, c=WHITE, bold=False, al=PP_ALIGN.LEFT, fontname='Microsoft YaHei'):
    tx = s.shapes.add_textbox(l, t, w, h); tf = tx.text_frame; tf.word_wrap = True
    p = tf.paragraphs[0]; p.text = text; p.font.size = Pt(fs)
    p.font.color.rgb = c; p.font.bold = bold; p.font.name = fontname
    p.alignment = al; return tx

def mtb(s, l, t, w, h, lines, al=PP_ALIGN.LEFT):
    tx = s.shapes.add_textbox(l, t, w, h); tf = tx.text_frame; tf.word_wrap = True
    for i, ln in enumerate(lines):
        p = tf.paragraphs[0] if i == 0 else tf.add_paragraph()
        p.text = ln.get('text',''); p.font.size = Pt(ln.get('size',12))
        p.font.color.rgb = ln.get('color',GRAY); p.font.bold = ln.get('bold',False)
        p.font.name = 'Microsoft YaHei'; p.alignment = al
        if 'sp' in ln: p.space_after = Pt(ln['sp'])
    return tx

def dbox(s, l, t, w, h, text, fill=ACCENT, fs=10, bold=True, c=WHITE):
    sh = rect(s, l, t, w, h, fill=fill)
    tf = sh.text_frame; tf.word_wrap = True; p = tf.paragraphs[0]
    p.text = text; p.font.size = Pt(fs); p.font.color.rgb = c
    p.font.bold = bold; p.font.name = 'Microsoft YaHei'; p.alignment = PP_ALIGN.CENTER
    return sh

def line(s, x1, y1, x2, y2, c=BORDER, w=2):
    cn = s.shapes.add_connector(MSO_CONNECTOR_TYPE.STRAIGHT, x1, y1, x2, y2)
    cn.line.color.rgb = c; cn.line.width = Pt(w); return cn

def aline(s, l, t, w, c=ACCENT):
    sh = s.shapes.add_shape(MSO_SHAPE.RECTANGLE, l, t, w, Pt(3))
    sh.fill.solid(); sh.fill.fore_color.rgb = c; sh.line.fill.background()

def circ(s, l, t, sz, fill=ACCENT):
    sh = s.shapes.add_shape(MSO_SHAPE.OVAL, l, t, sz, sz)
    sh.fill.solid(); sh.fill.fore_color.rgb = fill; sh.line.fill.background()

def pn(s, n, total):
    tb(s, Inches(12.3), Inches(7.0), Inches(0.8), Inches(0.35), f"{n}/{total}", fs=9, c=GRAY, al=PP_ALIGN.RIGHT)

def header(s, title, n, total):
    plain_rect(s, Inches(0), Inches(0), Inches(13.333), Inches(0.05), fill=ACCENT)
    tb(s, Inches(0.8), Inches(0.25), Inches(10), Inches(0.55), title, fs=26, c=WHITE, bold=True)
    aline(s, Inches(0.8), Inches(0.85), Inches(2.0))
    pn(s, n, total)

def section_header(s, section_name, title, n, total, color=ACCENT):
    """Section divider: 'Part X' label + title"""
    plain_rect(s, Inches(0), Inches(0), Inches(13.333), Inches(0.05), fill=color)
    dbox(s, Inches(0.8), Inches(0.3), Inches(2.2), Inches(0.32), section_name, fill=color, fs=11)
    tb(s, Inches(0.8), Inches(0.7), Inches(10), Inches(0.55), title, fs=26, c=WHITE, bold=True)
    aline(s, Inches(0.8), Inches(1.25), Inches(2.0), c=color)
    pn(s, n, total)

def arrow_right(s, x, y, c=BORDER):
    a = s.shapes.add_shape(MSO_SHAPE.RIGHT_ARROW, x, y, Inches(0.18), Inches(0.22))
    a.fill.solid(); a.fill.fore_color.rgb = c; a.line.fill.background()

def add_img_fit(s, filename, l, t, max_w, max_h, border_color=None):
    fp = os.path.join(IMGS, filename)
    if not os.path.exists(fp):
        rect(s, l, t, max_w, max_h, fill=RGBColor(0x12,0x22,0x38), bc=BORDER)
        tb(s, l+Inches(0.3), t+max_h/2-Inches(0.15), max_w-Inches(0.6), Inches(0.3),
           "[ IMG NOT FOUND ]", fs=10, c=GRAY, al=PP_ALIGN.CENTER)
        return (max_w, max_h)
    with Image.open(fp) as img:
        iw, ih = img.size
    if iw == 0 or ih == 0: return (max_w, max_h)
    scale = min(float(max_w)/iw, float(max_h)/ih)
    pw, ph = int(iw*scale), int(ih*scale)
    cx, cy = l+(max_w-pw)/2, t+(max_h-ph)/2
    rect(s, l, t, max_w, max_h, fill=RGBColor(0x12,0x22,0x38), bc=border_color or BORDER)
    s.shapes.add_picture(fp, cx, cy, pw, ph)
    return (pw, ph)

def screenshot_page(title, desc, img_file, color, page_num, total, section_label=""):
    s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
    # Clean title at top with accent bar
    plain_rect(s, Inches(0), Inches(0), Inches(13.333), Inches(0.05), fill=color)
    dbox(s, Inches(0.5), Inches(0.2), Inches(3.5), Inches(0.38), title, fill=color, fs=13)
    # Subtitle description
    tb(s, Inches(4.2), Inches(0.22), Inches(8.5), Inches(0.35), desc, fs=12, c=GRAY)
    # Full-screen image area
    fp = os.path.join(IMGS, img_file)
    if not os.path.isfile(fp):
        rect(s, Inches(0.3), Inches(0.8), Inches(12.7), Inches(6.5), fill=RGBColor(0x12,0x22,0x38), bc=color)
        tb(s, Inches(0.3), Inches(3.5), Inches(12.7), Inches(0.5),
           "[ 截图: " + img_file + " ]", fs=16, c=GRAY, al=PP_ALIGN.CENTER)
    else:
        with Image.open(fp) as img:
            iw, ih = img.size
        max_w, max_h = Inches(12.7), Inches(6.4)
        scale = min(float(max_w)/iw, float(max_h)/ih)
        pw, ph = int(iw*scale), int(ih*scale)
        cx, cy = Inches(0.3)+(max_w-pw)/2, Inches(0.75)+(max_h-ph)/2
        rect(s, Inches(0.3), Inches(0.75), max_w, max_h, fill=RGBColor(0x12,0x22,0x38), bc=color)
        s.shapes.add_picture(fp, cx, cy, pw, ph)
    pn(s, page_num, total)

# ===== SCREENSHOT LIST (4 key pages only) =====
shots = [
    ("工作流管理列表", "卡片式工作流管理：创建、搜索、删除，一键执行 & 内置模板", "02-工作流管理页面截图.png", GREEN),
    ("可视化编辑器", "基于 XYFlow 的拖拽式工作流编辑器 + 节点配置面板", "03-工作流编辑页面.png", ORANGE),
    ("工作流执行与结果", "一键执行，SSE 实时推送节点状态，完成后展示输出与耗时", "04-执行页面.png", LIGHT_B),
    ("Agent MCP 工具调用", "Agent 节点绑定 MCP Server，工具发现 → LLM 决策 → 工具执行", "05-mcp节点执行.png", PURPLE),
]
TOTAL_SHOTS = len(shots)
DESIGN_PAGES = 6  # S1 cover + S2-6 design
TOTAL = DESIGN_PAGES + TOTAL_SHOTS + 1  # + thanks page

# ==================================================================
# S1: COVER
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s, DEEPER, DARK_BG)
plain_rect(s, Inches(0), Inches(0), Inches(13.333), Inches(0.06), fill=ACCENT)
plain_rect(s, Inches(1.5), Inches(1.3), Inches(0.04), Inches(2.8), fill=ACCENT)
dbox(s, Inches(1.8), Inches(1.3), Inches(1.8), Inches(0.38), "MCPFlow", fill=ACCENT, fs=14)
tb(s, Inches(1.8), Inches(2.15), Inches(10.8), Inches(0.8),
   "基于 MCP 协议的多智能体工作流编排平台的设计与实现", fs=24, c=WHITE, bold=True)
plain_rect(s, Inches(1.8), Inches(3.2), Inches(3.5), Inches(0.03), fill=ACCENT)
mtb(s, Inches(1.8), Inches(3.5), Inches(5.0), Inches(3.0), [
    {'text': '答辩人：郑银杰', 'size': 16, 'c': GRAY, 'sp': 10},
    {'text': '专  业：计算机科学与技术', 'size': 16, 'c': GRAY, 'sp': 10},
    {'text': '学  院：计算机学院（人工智能学院）', 'size': 16, 'c': GRAY, 'sp': 10},
    {'text': '指导教师：唐菀 教授', 'size': 16, 'c': GRAY, 'sp': 10},
    {'text': '答辩日期：2026年5月', 'size': 16, 'c': GRAY},
])
for i in range(3):
    plain_rect(s, Inches(10.0), Inches(1.3)+Inches(1.3*i), Inches(2.5), Inches(0.005), fill=RGBColor(0x1B,0x3A,0x5C))
circ(s, Inches(11.3), Inches(5.8), Inches(1.4), fill=RGBColor(0x0A,0x20,0x3E))
circ(s, Inches(12.0), Inches(6.3), Inches(0.7), fill=RGBColor(0x14,0x30,0x50))
plain_rect(s, Inches(0), Inches(7.44), Inches(13.333), Inches(0.06), fill=ACCENT)
pn(s, 1, TOTAL)

# ==================================================================
# S2: PART 1 - MCP PROTOCOL & WORKFLOW BACKGROUND
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
section_header(s, "Part 1  工程背景", "MCP 协议与工作流编排", 2, TOTAL, ACCENT)

# Left column - MCP protocol
rect(s, Inches(0.8), Inches(1.6), Inches(5.8), Inches(2.7), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(0.8), Inches(1.6), Inches(5.8), Inches(0.05), fill=ACCENT)
tb(s, Inches(1.1), Inches(1.8), Inches(5.0), Inches(0.4), "MCP (Model Context Protocol)", fs=16, c=ACCENT, bold=True)
mtb(s, Inches(1.1), Inches(2.3), Inches(5.3), Inches(1.8), [
    {'text': 'Anthropic 提出，标准化 AI-工具通信', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '核心抽象：Resources / Prompts / Tools', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '传输机制：Streamable HTTP + SSE fallback', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '服务端注册工具，客户端自动发现调用', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '本系统使用 mcp-go 库实现客户端', 'size': 13, 'c': GRAY},
])

# Right column - why workflow
rect(s, Inches(7.2), Inches(1.6), Inches(5.3), Inches(2.7), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(7.2), Inches(1.6), Inches(5.3), Inches(0.05), fill=GREEN)
tb(s, Inches(7.5), Inches(1.8), Inches(5.0), Inches(0.4), "多智能体工作流", fs=16, c=GREEN, bold=True)
mtb(s, Inches(7.5), Inches(2.3), Inches(4.8), Inches(1.8), [
    {'text': 'LLM + Agent 快速发展，复杂任务需拆解', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '单次对话无法完成：需多步、多工具', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': 'DAG 工作流：有向无环，拓扑排序执行', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '条件分支实现动态流程跳转', 'size': 13, 'c': GRAY, 'sp': 12},
    {'text': '节点间通过 Context 共享数据', 'size': 13, 'c': GRAY},
])

# Bottom - project objectives
rect(s, Inches(0.8), Inches(4.6), Inches(11.7), Inches(2.6), fill=CARD_BG, bc=BORDER)
tb(s, Inches(1.3), Inches(4.8), Inches(6.0), Inches(0.35), "项目目标", fs=16, c=ACCENT, bold=True)
goals = [("低代码编排", "可视化拖拽，降低 AI 流程构建门槛", GREEN),
         ("MCP 协议集成", "标准化工具体接入，平台无关", ACCENT)]
for i, (title, desc, color) in enumerate(goals):
    x = Inches(1.5)+Inches(5.5*i)
    dbox(s, x, Inches(5.3), Inches(2.0), Inches(0.42), title, fill=color, fs=14)
    tb(s, x+Inches(2.2), Inches(5.32), Inches(2.8), Inches(0.4), desc, fs=12, c=GRAY)
feats = [("工作流可视化编辑器", "拖拽节点、连线、配置、保存", "基于 XYFlow + React"),
         ("DAG 工作流执行引擎", "拓扑排序 + 条件分支 + 变量传递", "Go + expr-lang + Goja"),
         ("Agent + MCP 工具调用", "工具发现 + Function Calling 循环", "mcp-go + OpenAI API"),
         ("容器化一键部署", "Docker Compose 启动全栈", "Docker + Nginx")]
for i, (title, desc, tech) in enumerate(feats):
    r, c = i//2, i%2; x = Inches(1.5)+Inches(6.0*c); y = Inches(5.9)+Inches(0.55*r)
    tb(s, x, y, Inches(5.5), Inches(0.22), title, fs=12, c=ACCENT, bold=True)
    tb(s, x, y+Inches(0.25), Inches(2.8), Inches(0.22), desc, fs=10, c=GRAY)
    tb(s, x+Inches(2.9), y+Inches(0.25), Inches(2.5), Inches(0.22), "["+tech+"]", fs=9, c=BORDER)

# ==================================================================
# S3: PART 2 - OVERALL ARCHITECTURE
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
section_header(s, "Part 2  设计思路", "系统总体架构设计", 3, TOTAL, GREEN)

# Four-layer architecture
layers = [
    ("Frontend", "React + XYFlow + Ant Design", Inches(2.55), ACCENT, [
        "Workflow Editor  |  XYFlow 拖拽编排",
        "Page UI  |  Ant Design 组件",
        "API Client  |  Axios 请求封装",
        "SSE Monitor  |  EventSource 监听",
    ]),
    ("Backend", "Go + Gin + DAG Engine", Inches(3.45), GREEN, [
        "REST API  |  工作流/Agent/MCP CRUD",
        "DAG Engine  |  拓扑排序 + 节点调度",
        "MCP Integration  |  工具发现与调用",
        "SSE Stream  |  实时推送执行状态",
    ]),
    ("Database", "MongoDB", Inches(4.35), LIGHT_B, [
        "Workflows Collection  |  工作流定义",
        "Executions Collection  |  执行记录",
        "MCP Servers  |  MCP 连接信息",
        "LLM Providers  |  模型配置",
    ]),
    ("External", "LLM Provider + MCP Server", Inches(5.25), ORANGE, [
        "LLM Provider  |  OpenAI-compatible API",
        "MCP Server  |  Streamable HTTP / SSE",
    ]),
]
for name, subtitle, y, color, items in layers:
    rect(s, Inches(0.8), y, Inches(11.7), Inches(0.85), fill=CARD_BG, bc=color)
    tb(s, Inches(1.2), y+Inches(0.05), Inches(3.0), Inches(0.3), name, fs=14, c=color, bold=True)
    tb(s, Inches(1.2), y+Inches(0.35), Inches(3.5), Inches(0.3), subtitle, fs=11, c=GRAY)
    for j, item in enumerate(items):
        tb(s, Inches(5.5)+Inches(3.3*(j%2)), y+Inches(0.1)+Inches(0.28*(j//2)), Inches(3.0), Inches(0.25), item, fs=9, c=GRAY)
    if name != "External":
        line(s, Inches(6.5), y+Inches(0.85), Inches(6.5), y+Inches(0.9), c=color, w=2)

# Right side - tech stack summary
for i, (name, color) in enumerate([("Go + Gin", GREEN), ("React + TS", ACCENT), ("MongoDB", LIGHT_B), ("Docker", ORANGE)]):
    dbox(s, Inches(9.6), Inches(6.15)+Inches(0.3*i), Inches(2.5), Inches(0.26), name, fill=color, fs=9)

# ==================================================================
# S4: NODE TYPES + TEMPLATE VARIABLE (no MongoDB doc structure)
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
section_header(s, "Part 2  设计思路", "核心设计：节点类型 & 模板变量机制", 4, TOTAL, ORANGE)

# --- TOP: Node Types as full-width 2x4 card grid ---
rect(s, Inches(0.8), Inches(1.6), Inches(11.7), Inches(3.3), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(0.8), Inches(1.6), Inches(11.7), Inches(0.05), fill=ORANGE)
tb(s, Inches(1.1), Inches(1.75), Inches(10.0), Inches(0.35), "8 种节点类型 — 各司其职", fs=16, c=ORANGE, bold=True)

node_types = [
    ("Start", "入口参数定义", GREEN, Inches(1.1), Inches(2.25)),
    ("LLM", "模型+提示词+温度", ACCENT, Inches(4.0), Inches(2.25)),
    ("Agent", "MCP服务+工具循环", ORANGE, Inches(7.0), Inches(2.25)),
    ("Condition", "表达式条件分支", PURPLE, Inches(10.0), Inches(2.25)),
    ("End", "结果输出展示", GREEN, Inches(1.1), Inches(3.25)),
    ("Code", "JS沙箱执行代码", SALMON, Inches(4.0), Inches(3.25)),
    ("HTTP", "REST API调用", LIGHT_B, Inches(7.0), Inches(3.25)),
    ("Email", "SMTP邮件发送", PURPLE, Inches(10.0), Inches(3.25)),
]
for name, desc, color, x, y in node_types:
    dbox(s, x, y, Inches(1.3), Inches(0.35), name, fill=color, fs=12)
    rect(s, x, y+Inches(0.42), Inches(2.5), Inches(0.55), fill=DARK_CARD, bc=color)
    tb(s, x+Inches(0.1), y+Inches(0.45), Inches(2.3), Inches(0.5), desc, fs=10, c=GRAY)

# --- BOTTOM: Context & Template Variable (visual flow diagram) ---
rect(s, Inches(0.8), Inches(5.2), Inches(11.7), Inches(2.0), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(0.8), Inches(5.2), Inches(11.7), Inches(0.05), fill=ORANGE)
tb(s, Inches(1.3), Inches(5.35), Inches(6.0), Inches(0.35), "执行上下文 & 模板变量机制", fs=16, c=ORANGE, bold=True)

# Visual flow: 5 colored blocks with arrows
flow_steps = [
    (Inches(1.3), "上游节点\n执行", ACCENT),
    (Inches(3.4), "输出写入\nContext KV", GREEN),
    (Inches(5.5), "模板变量\n引用解析", PURPLE),
    (Inches(7.6), "运行时\n替换注入", SALMON),
    (Inches(9.7), "下游节点\n使用结果", ACCENT),
]
for x, text, color in flow_steps:
    dbox(s, x, Inches(5.85), Inches(1.5), Inches(0.55), text, fill=color, fs=10)
for i in range(len(flow_steps)-1):
    ax = flow_steps[i][0] + Inches(1.55)
    arrow_right(s, ax, Inches(5.98), c=BORDER)

tb(s, Inches(1.3), Inches(6.6), Inches(11.0), Inches(0.35),
   "支持对象嵌套访问、数组索引和条件布尔表达式 — 实现节点间数据共享与动态流程控制", fs=12, c=GRAY)

# ==================================================================
# S5: WORKFLOW ENGINE DESIGN (NO CODE - pure visuals)
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
section_header(s, "Part 2  设计思路", "核心设计：DAG 工作流引擎", 5, TOTAL, GREEN)

# Top: 6-step engine flow (2 rows x 3 cols) — kept as-is, already pure visual
steps_data = [
    ("1", "拓扑排序", "Kahn BFS" + chr(10) + "确定执行顺序", ACCENT),
    ("2", "初始化上下文", "创建 KV 存储" + chr(10) + "注入用户输入", GREEN),
    ("3", "节点调度", "遍历拓扑序列" + chr(10) + "路由到对应 Handler", ORANGE),
    ("4", "条件分支", "评估条件表达式" + chr(10) + "跳过 false 路径", PURPLE),
    ("5", "保存 & 记录", "写 Context" + chr(10) + "记录耗时 + 输出", LIGHT_B),
    ("6", "SSE 推送", "实时推送状态" + chr(10) + "前端展示进度", SALMON),
]
for i, (num, title, desc, color) in enumerate(steps_data):
    row, col = i//3, i%3; x = Inches(0.8)+Inches(4.0*col); y = Inches(1.6)+Inches(1.5*row)
    circ(s, x+Inches(1.5), y, Inches(0.32), fill=color)
    tb(s, x+Inches(1.55), y+Inches(0.03), Inches(0.22), Inches(0.26), num, fs=9, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
    rect(s, x, y+Inches(0.45), Inches(3.6), Inches(0.85), fill=CARD_BG, bc=color)
    tb(s, x+Inches(0.15), y+Inches(0.5), Inches(3.3), Inches(0.3), title, fs=13, c=color, bold=True)
    tb(s, x+Inches(0.15), y+Inches(0.82), Inches(3.3), Inches(0.4), desc, fs=10, c=GRAY)

# --- Bottom-Left: Strategy Pattern — 8 Handler cards ---
rect(s, Inches(0.8), Inches(4.75), Inches(5.8), Inches(2.45), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(0.8), Inches(4.75), Inches(5.8), Inches(0.05), fill=GREEN)
tb(s, Inches(1.1), Inches(4.9), Inches(5.0), Inches(0.35), "策略模式：节点类型 → Handler 路由", fs=14, c=GREEN, bold=True)

handler_types = [
    ("Start", GREEN), ("LLM", ACCENT), ("Agent", ORANGE), ("Condition", PURPLE),
    ("Code", PURPLE), ("HTTP", SALMON), ("Email", SALMON), ("End", GREEN),
]
for i, (name, color) in enumerate(handler_types):
    r, c = i//4, i%4; x = Inches(1.1)+Inches(1.35*c); y = Inches(5.25)+Inches(0.55*r)
    dbox(s, x, y, Inches(1.15), Inches(0.4), name, fill=color, fs=10)

# Center arrow
dbox(s, Inches(1.1), Inches(6.35), Inches(5.5), Inches(0.4),
     "引擎通过拓扑排序依次调用 → 根据节点类型路由到对应 Handler", fill=CARD_BG, fs=10, c=GRAY)

# --- Bottom-Right: Code Node execution flow ---
rect(s, Inches(7.2), Inches(4.75), Inches(5.3), Inches(2.45), fill=CARD_BG, bc=BORDER)
plain_rect(s, Inches(7.2), Inches(4.75), Inches(5.3), Inches(0.05), fill=ORANGE)
tb(s, Inches(7.5), Inches(4.9), Inches(4.8), Inches(0.35), "Code 节点：Goja JS 沙箱执行", fs=14, c=ORANGE, bold=True)

code_flow = [
    ("用户编写\nJS 逻辑", ACCENT),
    ("Goja\n纯Go沙箱", GREEN),
    ("读取\nContext", PURPLE),
    ("运算\n处理", SALMON),
    ("写回\nContext", ACCENT),
]
for i, (text, color) in enumerate(code_flow):
    x = Inches(7.4)+Inches(1.05*i); y = Inches(5.45)
    rect(s, x, y, Inches(0.9), Inches(0.7), fill=color, bc=None)
    tb(s, x+Inches(0.05), y+Inches(0.05), Inches(0.8), Inches(0.6), text, fs=8, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
    if i < len(code_flow)-1:
        arrow_right(s, x+Inches(0.92), y+Inches(0.25), c=BORDER)

mtb(s, Inches(7.5), Inches(6.3), Inches(4.8), Inches(0.8), [
    {'text': '纯 Go 实现的 JS 虚拟机，无 OS 访问权限', 'size': 11, 'c': GRAY, 'sp': 6},
    {'text': '限制最大执行时间，保证沙箱安全', 'size': 11, 'c': GRAY, 'sp': 6},
    {'text': '通过 ctx.set() 写入结果供下游节点引用', 'size': 11, 'c': GRAY},
])

# ==================================================================
# S6: MCP INTEGRATION DESIGN
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s)
section_header(s, "Part 2  设计思路", "核心设计：Agent 与 MCP 协议集成", 6, TOTAL, PURPLE)

# Top: 7-step flow
tb(s, Inches(0.8), Inches(1.5), Inches(8.0), Inches(0.4), "Agent Node Execution (7-step Flow)", fs=18, c=PURPLE, bold=True)
pw = Inches(1.4); ph = Inches(1.0); ag = Inches(0.18)
fsteps = [("Connect" + chr(10) + "MCP Server", ACCENT), ("Discover" + chr(10) + "tools/list", GREEN),
          ("Function" + chr(10) + "Calling", PURPLE), ("LLM" + chr(10) + "Decision", ORANGE),
          ("Execute" + chr(10) + "tools/call", SALMON), ("Return" + chr(10) + "Result", LIGHT_B),
          ("Final" + chr(10) + "Response", ACCENT)]
nsteps = len(fsteps); tw = nsteps*(pw+ag)-ag; sx = (Inches(13.333)-tw)/2
for i, (text, color) in enumerate(fsteps):
    x = sx+i*(pw+ag); y = Inches(2.05)
    circ(s, x+Inches(0.45), y-Inches(0.22), Inches(0.22), fill=color)
    tb(s, x+Inches(0.47), y-Inches(0.2), Inches(0.18), Inches(0.2), str(i+1), fs=7, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
    rect(s, x, y+Inches(0.08), pw, Inches(0.9), fill=CARD_BG, bc=color)
    tb(s, x+Inches(0.05), y+Inches(0.18), pw-Inches(0.1), Inches(0.7), text, fs=9, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
    if i < nsteps-1: arrow_right(s, x+pw+Inches(0.02), y+Inches(0.4), c=color)

# Bottom: Three detail boxes
y3 = Inches(3.35)
# Box 1: Tool-Use Loop
rect(s, Inches(0.8), y3, Inches(3.8), Inches(1.6), fill=CARD_BG, bc=BORDER)
tb(s, Inches(1.0), y3+Inches(0.1), Inches(3.5), Inches(0.3), "Tool-Use Loop", fs=14, c=PURPLE, bold=True)
for i, item in enumerate(["1. LLM returns tool_calls", "2. callTool via MCP client", "3. Result fed back to LLM", "4. Loop until text response"]):
    tb(s, Inches(1.0), y3+Inches(0.45)+Inches(0.26*i), Inches(3.5), Inches(0.25), item, fs=10, c=GRAY)

# Box 2: Transport compatibility
rect(s, Inches(4.9), y3, Inches(3.5), Inches(1.6), fill=CARD_BG, bc=BORDER)
tb(s, Inches(5.1), y3+Inches(0.1), Inches(3.2), Inches(0.3), "MCP Transport", fs=14, c=LIGHT_B, bold=True)
for i, item in enumerate(["Streamable HTTP (primary)", "SSE fallback (legacy)", "Auto-detect on connect", "Bidirectional streaming"]):
    tb(s, Inches(5.1), y3+Inches(0.45)+Inches(0.26*i), Inches(3.2), Inches(0.25), item, fs=10, c=GRAY)

# Box 3: Agent monitoring
rect(s, Inches(8.7), y3, Inches(3.9), Inches(1.6), fill=CARD_BG, bc=BORDER)
tb(s, Inches(8.9), y3+Inches(0.1), Inches(3.6), Inches(0.3), "Agent Monitoring", fs=14, c=SALMON, bold=True)
for i, item in enumerate(["Tool Name + Params + Result", "Per-step duration logging", "SSE real-time push", "Full execution trace"]):
    tb(s, Inches(8.9), y3+Inches(0.45)+Inches(0.26*i), Inches(3.6), Inches(0.25), item, fs=10, c=GRAY)

# Bottom section: Agent implementation — 3 columns visual cards (no code)
rect(s, Inches(0.8), Inches(5.25), Inches(11.7), Inches(2.0), fill=CARD_BG, bc=BORDER)
tb(s, Inches(1.1), Inches(5.35), Inches(6.0), Inches(0.35), "Agent Handler 实现要点", fs=14, c=PURPLE, bold=True)

# Column 1: Setup
rect(s, Inches(1.0), Inches(5.8), Inches(3.5), Inches(1.3), fill=DARK_CARD, bc=BORDER)
tb(s, Inches(1.2), Inches(5.9), Inches(3.0), Inches(0.3), "阶段一：准备", fs=13, c=ACCENT, bold=True)
for j, item in enumerate(["连接 MCP Server", "并行调用 tools/list", "聚合所有工具列表", "转为 OpenAI Tool 格式"]):
    circ(s, Inches(1.2), Inches(6.3)+Inches(0.2*j), Inches(0.12), fill=ACCENT)
    tb(s, Inches(1.4), Inches(6.28)+Inches(0.2*j), Inches(3.0), Inches(0.2), item, fs=10, c=GRAY)

# Column 2: Core loop
rect(s, Inches(4.8), Inches(5.8), Inches(3.5), Inches(1.3), fill=DARK_CARD, bc=BORDER)
tb(s, Inches(5.0), Inches(5.9), Inches(3.0), Inches(0.3), "阶段二：工具调用循环", fs=13, c=GREEN, bold=True)
for j, item in enumerate(["解析上下文模板变量", "构建 Prompt → LLM", "LLM 返回 tool_calls", "调用 MCP tools/call"]):
    circ(s, Inches(5.0), Inches(6.3)+Inches(0.2*j), Inches(0.12), fill=GREEN)
    tb(s, Inches(5.2), Inches(6.28)+Inches(0.2*j), Inches(3.0), Inches(0.2), item, fs=10, c=GRAY)

# Column 3: Output & monitor
rect(s, Inches(8.6), Inches(5.8), Inches(3.5), Inches(1.3), fill=DARK_CARD, bc=BORDER)
tb(s, Inches(8.8), Inches(5.9), Inches(3.0), Inches(0.3), "阶段三：监控与输出", fs=13, c=SALMON, bold=True)
for j, item in enumerate(["工具返回 → 回传 LLM", "SSE 实时推送步骤", "记录耗时与结果", "LLM 生成最终回答"]):
    circ(s, Inches(8.8), Inches(6.3)+Inches(0.2*j), Inches(0.12), fill=SALMON)
    tb(s, Inches(9.0), Inches(6.28)+Inches(0.2*j), Inches(3.0), Inches(0.2), item, fs=10, c=GRAY)

# ==================================================================
# S7-S22: SCREENSHOT PAGES (Part 3)
# ==================================================================
for idx, (title, desc, img_file, color) in enumerate(shots):
    page_num = DESIGN_PAGES + 1 + idx
    section = "Part 3  系统演示" if idx == 0 else ""
    screenshot_page(title, desc, img_file, color, page_num, TOTAL, section_label=section)

# ==================================================================
# THANKS
# ==================================================================
s = prs.slides.add_slide(prs.slide_layouts[6]); bg(s, DEEPER, DARK_BG)
plain_rect(s, Inches(0), Inches(0), Inches(13.333), Inches(0.05), fill=ACCENT)
plain_rect(s, Inches(0), Inches(7.44), Inches(13.333), Inches(0.06), fill=ACCENT)
tb(s, Inches(0), Inches(1.5), Inches(13.333), Inches(1.2), "致  谢", fs=48, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
aline(s, Inches(5.5), Inches(2.8), Inches(2.333))
mtb(s, Inches(2.0), Inches(3.2), Inches(9.333), Inches(3.0), [
    {'text': '感谢唐菀教授在选题、设计和论文撰写中的悉心指导', 'size': 18, 'c': GRAY, 'sp': 22},
    {'text': '感谢计算机学院（人工智能学院）老师们的培养', 'size': 18, 'c': GRAY, 'sp': 22},
    {'text': '感谢同学们和家人的支持', 'size': 18, 'c': GRAY, 'sp': 22},
], al=PP_ALIGN.CENTER)
plain_rect(s, Inches(5.0), Inches(5.8), Inches(3.333), Inches(0.02), fill=ACCENT)
tb(s, Inches(0), Inches(6.0), Inches(13.333), Inches(0.6), "恳请各位老师批评指正", fs=20, c=WHITE, bold=True, al=PP_ALIGN.CENTER)
pn(s, TOTAL, TOTAL)

# ===== SAVE =====
out = os.path.join(BASE, 'MCPFlow_毕业答辩.pptx')
prs.save(out)
print(f"OK: {out}")
print(f"Slides: {len(prs.slides)}")
