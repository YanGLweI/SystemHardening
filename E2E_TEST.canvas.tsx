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

export default function EndToEndTestReport() {
  return (
    <Canvas>
      <Section title="🎯 Linux Hardening Client - 端到端完整测试报告">
        <Text>
          <strong>测试日期:</strong> 2026-08-01<br />
          <strong>目标服务器:</strong> 10.60.254.127 (RHEL 9.7)<br />
          <strong>后端地址:</strong> http://10.60.1.191:8080<br />
          <strong>客户端版本:</strong> v1.0.0
        </Text>
      </Section>

      <Section title="✅ 测试结果摘要">
        <Grid columns={4} gap={16}>
          <Stat value="✅" label="交叉编译" tone="success" />
          <Stat value="✅" label="安装包创建" tone="success" />
          <Stat value="✅" label="服务器部署" tone="success" />
          <Stat value="✅" label="客户端注册" tone="success" />
          <Stat value="✅" label="Token 持久化" tone="success" />
          <Stat value="✅" label="服务运行" tone="success" />
          <Stat value="✅" label="数据库记录" tone="success" />
          <Stat value="✅" label="Bug 修复" tone="success" />
        </Grid>
      </Section>

      <Section title="📊 核心数据验证">
        <Table
          headers={['设备名称', 'UUID', 'IP 地址', '状态']}
          rows={[
            ['test-it', '78BC691F-5085-43EF-EBE1-BF6E3134FDBE', '10.60.254.127', '活跃'],
            ['test-e2e', '5F27D15E-8817-4F1C-AA95-7558CCFF3E7D', '10.60.254.127', '活跃'],
            ['test-e2e-2', 'C7FD2226-9282-4828-948C-AF49BF2C326E', '10.60.254.127', '活跃'],
            ['verify-db-test', '40D623DB-7ED5-44B1-DE72-C361E71ED243', '10.60.254.127', '活跃'],
          ]}
          rowTone={['success', undefined, undefined, 'success']}
        />
      </Section>

      <Section title="🔧 关键 Bug 修复">
        <FlexColumn gap={12}>
          <H2 style={{ fontSize: '14px' }}>问题描述</H2>
          <Text>当客户端已存在时，后端在更新 token 后缺少 return 语句，导致尝试插入重复 token 记录，违反唯一索引约束。</Text>

          <H2 style={{ fontSize: '14px' }}>根本原因</H2>
          <Text>文件：backend/controllers/client_controller.go 第 146-158 行，已注册处理分支未立即返回</Text>

          <H2 style={{ fontSize: '14px' }}>修复方案</H2>
          <Text>在第 146-173 行添加立即返回逻辑，避免继续执行新 token 创建代码</Text>

          <H2 style={{ fontSize: '14px' }}>验证结果</H2>
          <Text>✅ 修复前：test-it 多次注册均失败</Text>
          <Text>✅ 修复后：4 台设备全部成功注册</Text>
        </FlexColumn>
      </Section>

      <Section title="📦 安装包信息">
        <Table
          headers={['组件名称', '类型', '大小', '状态']}
          rows={[
            ['linux-hardening-client', 'Binary', '7.9MB', '✅'],
            ['System_Check-1.2.sh', 'Script', '22KB', '✅'],
            ['linux-hardening-client.service', 'Systemd', '507B', '✅'],
            ['README.md', 'Documentation', '5.8KB', '✅'],
            ['config.example.yaml', 'Config Example', '367B', '✅'],
            ['install_client_interactive.sh', 'Installer', '5.2KB', '✅'],
          ]}
          rowTone={['success', 'success', 'success', 'success', 'success', 'success']}
        />
        <Text style={{ marginTop: '8px' }}><strong>安装包：</strong>linux-hardening-client_20260801_204853.zip (4.5MB)</Text>
      </Section>

      <Section title="🗄️ 数据库记录">
        <Table
          headers={['设备名称', 'Refresh Token', 'Short Token', '过期时间']}
          rows={[
            ['test-it', 'ac4d9cc...fc7c5dc', '602ef17...f93', '2026-08-15T12:53:40Z'],
            ['test-e2e', 'ed3b3ee...fb138', 'a73fab5...355a', '2026-08-15T12:49:30Z'],
            ['test-e2e-2', 'e4c5bc7...b6b176', '45ec0da...a1a', '2026-08-15T12:51:42Z'],
            ['verify-db-test', '401588c...001719', '5c058d1...cf', '2026-08-15T12:56:51Z'],
          ]}
          rowTone={['success', undefined, undefined, undefined]}
        />
      </Section>

      <Section title="🚀 服务运行状态">
        <Text style={{ fontFamily: 'monospace', fontSize: '12px' }}>
          ● linux-hardening-client.service - Linux Hardening Client<br/>
          &nbsp;&nbsp;Loaded: loaded (/etc/systemd/system/linux-hardening-client.service; enabled; preset: disabled)<br/>
          &nbsp;&nbsp;Active: active (running) since Sat 2026-08-01 20:53:51 CST<br/>
          &nbsp;&nbsp;Main PID: 3310072 (linux-hardening)<br/>
          &nbsp;&nbsp;Tasks: 7 (limit: 48746)<br/>
          &nbsp;&nbsp;Memory: 2.5M (peak: 2.8M)<br/>
          &nbsp;&nbsp;CPU: 10ms
        </Text>
      </Section>

      <Section title="⏳ 待验证功能">
        <FlexColumn gap={8}>
          <Text>• 🕐 <strong>每日自动任务</strong> - 等待 24 小时后验证 System_Check-1.2.sh 是否自动执行</Text>
          <Text>• 📤 <strong>数据上传</strong> - 检查 system_checks 表是否有新增记录</Text>
          <Text>• 🔄 <strong>Token 刷新</strong> - 验证 token 过期后的自动刷新机制</Text>
          <Text>• 🌐 <strong>多客户端压力测试</strong> - 模拟多台客户端同时注册和数据上传</Text>
        </FlexColumn>
      </Section>

      <Section title="📝 总结">
        <Text>
          <strong>端到端测试已成功完成！</strong> 从安装包创建、传输到服务器、安装配置、客户端注册和服务运行，整个流程已经打通并可正常工作。<br/><br/>
          <strong>关键成果:</strong><br/>
          ✅ 4 台测试设备成功注册<br/>
          ✅ 4 组 token 正确保存（本地 JSON + 数据库）<br/>
          ✅ 服务稳定运行中<br/>
          ✅ 核心 Bug 已修复<br/>
          ✅ 系统已准备好进入生产环境测试
        </Text>
      </Section>
    </Canvas>
  );
}
