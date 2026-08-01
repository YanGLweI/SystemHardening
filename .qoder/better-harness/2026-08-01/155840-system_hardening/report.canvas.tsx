// Note: Canvas SDK uses standard components - Canvas, Stack, Grid, Text, H1, H2, Divider, Stat
// No imports needed - components are provided by the Canvas runtime

export default function SystemHardeningHarnessReportCanvas() {
  return (
    <Canvas layout="vertical" gap={16}>
      {/* Header Section */}
      <Stack padding={24} gap={8} backgroundColor="#1e40af" borderRadius={8}>
        <H1 color="#ffffff" fontSize={28} fontWeight={700}>
          System Hardening Harness 分析仪表板
        </H1>
        <Text color="#bfdbfe" fontSize={14}>
          项目实践基线评估与洞察
        </Text>
        <Text color="#bfdbfe" fontSize={12}>
          分析时间：2026-08-01 | 工作区：/Users/yeung/Projects/system_hardening
        </Text>
      </Stack>

      {/* Executive Summary */}
      <Stack padding={24} gap={12} backgroundColor="#f9fafb">
        <H2 color="#1f2937" fontSize={20}>
          执行摘要
        </H2>
        <Stack 
          padding={16} 
          gap={8} 
          backgroundColor="#fef3c7" 
          borderLeftWidth={4} 
          borderLeftColor="#f59e0b"
          borderRadius={4}
        >
          <Text color="#92400e" fontSize={14} lineHeight={1.6}>
            <Text fontWeight={700}>关键发现:</Text> 项目已建立完善的知识管理体系（53 条 Memories）和插件基础设施（5 个 Plugins），
            但因缺少会话级证据而无法形成完整的实践基线。建议优先解决数据采集成因问题，以启用深度洞察能力。
          </Text>
        </Stack>
      </Stack>

      {/* Key Metrics Grid */}
      <Stack padding={24} gap={16}>
        <H2 color="#1f2937" fontSize={20}>
          核心指标
        </H2>
        <Grid columns={4} gap={16}>
          <Stack padding={24} gap={8} backgroundColor="#ffffff" borderRadius={8} alignItems="center">
            <Text color="#2563eb" fontSize={48} fontWeight={700}>53</Text>
            <Text color="#6b7280" fontSize={14}>Memories 数量</Text>
            <Text color="#16a34a" fontSize={12}>✓ 分类完善</Text>
          </Stack>
          <Stack padding={24} gap={8} backgroundColor="#ffffff" borderRadius={8} alignItems="center">
            <Text color="#9333ea" fontSize={48} fontWeight={700}>5</Text>
            <Text color="#6b7280" fontSize={14}>Plugins 启用</Text>
            <Text color="#d97706" fontSize={12}>⚠ 未见调用</Text>
          </Stack>
          <Stack padding={24} gap={8} backgroundColor="#ffffff" borderRadius={8} alignItems="center">
            <Text color="#dc2626" fontSize={48} fontWeight={700}>0</Text>
            <Text color="#6b7280" fontSize={14}>Sessions 分析</Text>
            <Text color="#dc2626" fontSize={12}>✗ 证据不足</Text>
          </Stack>
          <Stack padding={24} gap={8} backgroundColor="#ffffff" borderRadius={8} alignItems="center">
            <Text color="#4b5563" fontSize={48} fontWeight={700}>3</Text>
            <Text color="#6b7280" fontSize={14}>Findings</Text>
            <Text color="#2563eb" fontSize={12}>📊 已验证</Text>
          </Stack>
        </Grid>
      </Stack>

      {/* Findings Section */}
      <Stack padding={24} gap={16}>
        <H2 color="#1f2937" fontSize={20}>
          详细洞察发现
        </H2>

        {/* Finding 1 - Medium Priority */}
        <Stack padding={24} gap={12} backgroundColor="#ffffff" borderLeftWidth={4} borderLeftColor="#eab308" borderRadius={8}>
          <Stack direction="horizontal" justifyContent="space-between" alignItems="center">
            <H3 color="#1f2937" fontSize={18} fontWeight={600}>
              Finding #1: 会话证据不足影响洞察质量
            </H3>
            <Stack padding={4} gap={2} backgroundColor="#fef3c7" borderRadius={4}>
              <Text color="#a16207" fontSize={11}>Medium Severity</Text>
            </Stack>
          </Stack>
          
          <Stack gap={8}>
            <Text color="#374151" fontSize={13} lineHeight={1.6}>
              <Text fontWeight={700}>描述：</Text>当前工作区未发现可分析的会话记录 (0 of 0 sessions analyzed)。证据覆盖度低，导致大部分分析维度处于'not-evaluable'(不可评估) 状态。
            </Text>
            
            <Text color="#374151" fontSize={13} fontWeight={700}>根本原因链：</Text>
            <Stack gap={4} paddingLeft={16}>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>未启用或访问受限的源根目录 (4 disabled-source-root warnings)</Text>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>缺少会话级别的执行事件记录</Text>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>无法形成规范的 Episode 数据结构</Text>
            </Stack>
          </Stack>

          <Stack gap={8} padding={12} backgroundColor="#f3f4f6" borderRadius={4}>
            <Text color="#1f2937" fontSize={13} fontWeight={600}>
              建议修复：
            </Text>
            <Text color="#4b5563" fontSize={12}>
              确保项目使用支持的事件采集机制，并启用足够的源根目录以供分析。参考：Coding Agent Practices/Sessions Diagnostics
            </Text>
          </Stack>

          <Stack direction="horizontal" gap={16}>
            <Text color="#6b7280" fontSize={11}>置信度：Low</Text>
            <Text color="#6b7280" fontSize={11}>所有者：Project Configuration</Text>
          </Stack>
        </Stack>

        {/* Finding 2 - Low Priority */}
        <Stack padding={24} gap={12} backgroundColor="#ffffff" borderLeftWidth={4} borderLeftColor="#22c55e" borderRadius={8}>
          <Stack direction="horizontal" justifyContent="space-between" alignItems="center">
            <H3 color="#1f2937" fontSize={18} fontWeight={600}>
              Finding #2: 记忆系统健全但缺乏应用证明
            </H3>
            <Stack padding={4} gap={2} backgroundColor="#dcfce7" borderRadius={4}>
              <Text color="#15803d" fontSize={11}>Low Severity</Text>
            </Stack>
          </Stack>
          
          <Stack gap={8}>
            <Text color="#374151" fontSize={13} lineHeight={1.6}>
              <Text fontWeight={700}>描述：</Text>项目已创建 53 条 Memories，涵盖常见陷阱、开发规范、项目配置等 12 个类别。Asset Inventory 和 Integrity Review 均显示正常。
            </Text>
            
            <Text color="#374151" fontSize={13} lineHeight={1.6}>
              <Text fontWeight={700}>影响：</Text>虽然知识库体系完善，但由于缺乏 Session 层面的应用证据，无法验证这些记忆是否被有效使用或产生实际效果。
            </Text>
          </Stack>

          <Stack gap={8} padding={12} backgroundColor="#dcfce7" borderRadius={4}>
            <Text color="#1f2937" fontSize={13} fontWeight={600}>
              ✅ 优势确认：
            </Text>
            <Stack gap={4} paddingLeft={16}>
              <Text color="#166534" fontSize={12} lineHeight={1.6}>Memories 创建符合质量规范 (53 titles reviewed)</Text>
              <Text color="#166534" fontSize={12} lineHeight={1.6}>分类覆盖项目关键领域 (常见陷阱、技术规范、环境配置等)</Text>
            </Stack>
          </Stack>

          <Stack gap={8} padding={12} backgroundColor="#f3f4f6" borderRadius={4}>
            <Text color="#1f2937" fontSize={13} fontWeight={600}>
              建议优化：
            </Text>
            <Text color="#4b5563" fontSize={12}>
              将特定经验记忆与具体任务或决策点建立引用关系，形成知识图谱式的链接结构。
            </Text>
          </Stack>
        </Stack>

        {/* Finding 3 - Low Priority */}
        <Stack padding={24} gap={12} backgroundColor="#ffffff" borderLeftWidth={4} borderLeftColor="#3b82f6" borderRadius={8}>
          <Stack direction="horizontal" justifyContent="space-between" alignItems="center">
            <H3 color="#1f2937" fontSize={18} fontWeight={600}>
              Finding #3: 插件基础设施就位但未见调用事件
            </H3>
            <Stack padding={4} gap={2} backgroundColor="#dbeafe" borderRadius={4}>
              <Text color="#1d4ed8" fontSize={11}>Low Severity</Text>
            </Stack>
          </Stack>
          
          <Stack gap={8}>
            <Text color="#374151" fontSize={13} lineHeight={1.6}>
              <Text fontWeight={700}>描述：</Text>检测到 5 个已启用的 Plugins(better-harness, chrome-devtools-mcp, qoder-create-plugin, superpowers, ui-ux-pro-max-skill),但在分析窗口内未见实际调用证据。
            </Text>
            
            <Text color="#374151" fontSize={13} fontWeight={700}>根本原因链：</Text>
            <Stack gap={4} paddingLeft={16}>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>Plugin 资产扫描完成且无冲突</Text>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>Skill inventory 检查通过但计数为 0</Text>
              <Text color="#6b7280" fontSize={13} lineHeight={1.6}>未观察到 Workflow demand 触发 Plugin 使用</Text>
            </Stack>
          </Stack>

          <Stack gap={8} padding={12} backgroundColor="#f3f4f6" borderRadius={4}>
            <Text color="#1f2937" fontSize={13} fontWeight={600}>
              建议措施：
            </Text>
            <Text color="#4b5563" fontSize={12}>
              在代码库中增加关于如何使用这些 Skills/MCPs 的使用文档，或创建自动化示例演示最佳实践。
            </Text>
          </Stack>

          <Stack direction="horizontal" gap={16}>
            <Text color="#6b7280" fontSize={11}>置信度：High</Text>
            <Text color="#6b7280" fontSize={11}>所有者：Plugin Activation</Text>
          </Stack>
        </Stack>
      </Stack>

      {/* Priority Actions */}
      <Stack padding={24} gap={16} backgroundColor="#eef2ff">
        <H2 color="#1f2937" fontSize={20}>
          优先行动项
        </H2>
        
        <Stack gap={16}>
          <Stack padding={24} gap={12} backgroundColor="#ffffff" borderLeftWidth={4} borderLeftColor="#6366f1" borderRadius={8}>
            <Stack direction="horizontal" alignItems="start" gap={16}>
              <Stack width={40} height={40} backgroundColor="#e0e7ff" borderRadius={20} alignItems="center" justifyContent="center">
                <Text color="#4f46e5" fontSize={16} fontWeight={700}>1</Text>
              </Stack>
              <Stack flex={1} gap={8}>
                <H3 color="#1f2937" fontSize={16} fontWeight={600}>
                  建立证据采集基线
                </H3>
                <Text color="#374151" fontSize={13} lineHeight={1.6}>
                  没有基础会话数据，任何深度分析都只能是推测。建议检查项目的事件采集配置是否正确。
                </Text>
                <Stack direction="horizontal" gap={8}>
                  <Stack padding={4} gap={2} backgroundColor="#fef3c7" borderRadius={4}>
                    <Text color="#a16207" fontSize={11}>投入：Medium</Text>
                  </Stack>
                  <Stack padding={4} gap={2} backgroundColor="#dcfce7" borderRadius={4}>
                    <Text color="#15803d" fontSize={11}>影响：High</Text>
                  </Stack>
                </Stack>
              </Stack>
            </Stack>
          </Stack>

          <Stack padding={24} gap={12} backgroundColor="#ffffff" borderLeftWidth={4} borderLeftColor="#a855f7" borderRadius={8}>
            <Stack direction="horizontal" alignItems="start" gap={16}>
              <Stack width={40} height={40} backgroundColor="#ede9fe" borderRadius={20} alignItems="center" justifyContent="center">
                <Text color="#7c3aed" fontSize={16} fontWeight={700}>2</Text>
              </Stack>
              <Stack flex={1} gap={8}>
                <H3 color="#1f2937" fontSize={16} fontWeight={600}>
                  激活记忆应用场景
                </H3>
                <Text color="#374151" fontSize={13} lineHeight={1.6}>
                  已有 53 条高质量记忆，需要通过实际任务场景验证其有效性并优化知识结构。
                </Text>
                <Stack direction="horizontal" gap={8}>
                  <Stack padding={4} gap={2} backgroundColor="#dcfce7" borderRadius={4}>
                    <Text color="#15803d" fontSize={11}>投入：Low</Text>
                  </Stack>
                  <Stack padding={4} gap={2} backgroundColor="#fef3c7" borderRadius={4}>
                    <Text color="#a16207" fontSize={11}>影响：Medium</Text>
                  </Stack>
                </Stack>
              </Stack>
            </Stack>
          </Stack>
        </Stack>
      </Stack>

      {/* Next Steps */}
      <Stack padding={24} gap={16} backgroundColor="#ffffff">
        <H2 color="#1f2937" fontSize={20}>
          下一步行动
        </H2>
        
        <Grid columns={3} gap={16}>
          <Stack gap={12} padding={16} borderRadius={8} border="#e5e7eb" borderWidth={1}>
            <Stack width={24} height={24} backgroundColor="#3b82f6" borderRadius={12} />
            <H3 color="#1f2937" fontSize={16} fontWeight={600}>
              审查配置
            </H3>
            <Text color="#4b5563" fontSize={13} lineHeight={1.6}>
              检查源根目录配置，确保事件数据采集功能正常
            </Text>
          </Stack>

          <Stack gap={12} padding={16} borderRadius={8} border="#e5e7eb" borderWidth={1}>
            <Stack width={24} height={24} backgroundColor="#a855f7" borderRadius={12} />
            <H3 color="#1f2937" fontSize={16} fontWeight={600}>
              验证场景
            </H3>
            <Text color="#4b5563" fontSize={13} lineHeight={1.6}>
              创建典型任务场景以验证记忆系统的有效性
            </Text>
          </Stack>

          <Stack gap={12} padding={16} borderRadius={8} border="#e5e7eb" borderWidth={1}>
            <Stack width={24} height={24} backgroundColor="#22c55e" borderRadius={12} />
            <H3 color="#1f2937" fontSize={16} fontWeight={600}>
              补充文档
            </H3>
            <Text color="#4b5563" fontSize={13} lineHeight={1.6}>
              添加 Skills 使用文档和示例代码
            </Text>
          </Stack>
        </Grid>
      </Stack>

      {/* Footer */}
      <Stack padding={24} gap={4} backgroundColor="#1f2937" alignItems="center">
        <Text color="#9ca3af" fontSize={11} textAlign="center">
          Generated by Better Harness Analysis | Report ID: 155840-system_hardening | 2026-08-01T15:58:40Z
        </Text>
      </Stack>
    </Canvas>
