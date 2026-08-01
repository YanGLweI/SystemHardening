import { 
  Canvas, 
  Divider, 
  FlexColumn, 
  Grid, 
  H1, 
  H2, 
  Section, 
  Stat, 
  Table, 
  Text 
} from 'qoder/canvas';

export default function InteractiveInstallerReport() {
  return (
    <Canvas>
      <Section title="🎯 交互式安装脚本 - 完整测试报告">
        <Text>
          <strong>测试目标:</strong> 验证改进后的 install_client_interactive.sh 能够自动获取系统信息<br/>
          <strong>关键特性:</strong> ✅ 自动设备名 | ✅ 自动 IP 地址 | ✅ 一键部署<br/>
          <strong>服务器:</strong> 10.60.254.127 (test-it) | <strong>后端:</strong> http://10.60.1.191:8080
        </Text>
      </Section>

      <Section title="✅ 核心改进点">
        <Grid columns={3} gap={16}>
          <Stat value="✅" label="自动获取 hostname" tone="success" />
          <Stat value="✅" label="自动获取 IP" tone="success" />
          <Stat value="✅" label="模式智能检测" tone="success" />
          <Stat value="✅" label="配置文件自动生成" tone="success" />
          <Stat value="✅" label="Token JSON 格式" tone="success" />
          <Stat value="✅" label="一键安装完成" tone="success" />
        </Grid>
      </Section>

      <Section title="📊 部署流程对比">
        <Table
          headers={['步骤', '之前（手动）', '现在（自动化）']}
          rows={[
            ['设备名配置', '手动创建 config.yaml<br>填写 test-20260801_210535', '自动从 hostname 获取 → test-it'],
            ['IP 地址配置', '手动输入 10.60.254.127', '自动从网卡获取 → 10.60.254.127'],
            ['Token 路径', '/data/tokens.db (SQLite)', '/data/tokens.json (纯 JSON)'],
            ['文件来源', '../bin/ (本地路径问题)', '当前目录自动探测'],
            ['运行环境', '只能本地开发', '服务器/本地双模式'],
            ['最终体验', '多步操作，容易出错', '一行命令完成部署'],
          ]}
          rowTone={['success', undefined, 'success', undefined, 'success', undefined]}
        />
      </Section>

      <Section title="⚙️ 技术实现">
        <FlexColumn gap={12}>
          <H2 style={{ fontSize: '14px' }}>模式检测逻辑</H2>
          <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
if [ -f "${SCRIPT_DIR}/linux-hardening-client" ] && \
   [ -f "${SCRIPT_DIR}/System_Check-1.2.sh" ]; then
    echo "✅ Running in SERVER MODE"
else
    echo "⚠️  Running in DEVELOPMENT MODE"
fi
          </Text>

          <H2 style={{ fontSize: '14px' }}>自动获取系统信息</H2>
          <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
LOCAL_HOSTNAME=$(hostname)
PRIMARY_IP=$(hostname -I | awk '{print $1}')
          </Text>

          <H2 style={{ fontSize: '14px' }}>配置生成示例</H2>
          <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
server_url: http://10.60.1.191:8080
local_db_path: /opt/linux-hardening-client/data/tokens.json
device_name: test-it              ← auto-detected
ip_address: 10.60.254.127         ← auto-detected
          </Text>
        </FlexColumn>
      </Section>

      <Section title="🗄️ 数据库状态">
        <Table
          headers={['字段', '值', '备注']}
          rows={[
            ['device_name', 'test-it', '自动获取的 hostname'],
            ['ip_address', '10.60.254.127', '自动获取的 IP'],
            ['client_uuid', '78BC691F-...', '后端分配的 UUID'],
            ['tokens.json', '已保存', 'short_token + refresh_token'],
          ]}
          rowTone={['success', undefined, undefined, undefined]}
        />
      </Section>

      <Section title="🚀 服务状态">
        <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
● linux-hardening-client.service - Linux Hardening Client<br/>
&nbsp;&nbsp;Active: active (running) since Sat 2026-08-01 21:35:35 CST<br/>
&nbsp;&nbsp;Main PID: 3313277 (linux-hardening)<br/>
&nbsp;&nbsp;Memory: 4.0M | CPU: 15ms
        </Text>
      </Section>

      <Section title="📈 测试结果">
        <Grid columns={4} gap={16}>
          <Stat value="100%" label="注册成功率" tone="success" />
          <Stat value="auto" label="设备名" tone="success" />
          <Stat value="auto" label="IP 地址" tone="success" />
          <Stat value="✓" label="服务运行" tone="success" />
        </Grid>
      </Section>

      <Section title="💡 使用示例">
        <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
# 上传安装包到服务器
scp linux-hardening-client*.zip root@10.60.254.127:/root/

# 解压并运行交互式安装脚本
cd /root && unzip *.zip && \
bash install_client_interactive.sh http://10.60.1.191:8080

# 查看日志
journalctl -u linux-hardening-client -f
        </Text>
      </Section>

      <Section title="📝 总结">
        <Text>
          <strong>交互式安装脚本已成功优化！</strong><br/><br/>
          主要成就：<br/>
          • ✅ 完全自动化设备识别（无需手动输入）<br/>
          • ✅ 智能模式检测（服务器/本地）<br/>
          • ✅ 正确配置文件生成<br/>
          • ✅ 一键部署体验<br/><br/>
          <strong>用户体验提升:</strong> 从复杂的手动配置简化为简单的一行命令，非常适合批量部署和 CI/CD 集成！
        </Text>
      </Section>
    </Canvas>
  );
}
