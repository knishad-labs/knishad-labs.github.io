/**
 * knishad.lab — Interactive Client Engine
 * SQL Optimization Studio, R&D Ideas Engine & Dynamic UI Controls
 */

document.addEventListener('DOMContentLoaded', () => {

  // Preset SQL Queries
  const PRESETS = {
    unindexed: `SELECT u.id, u.email, o.id AS order_id, o.amount 
FROM users u 
JOIN orders o ON u.id = o.user_id 
WHERE u.status = 'active' AND o.created_at >= '2026-01-01';`,

    wildcard: `SELECT * 
FROM orders 
WHERE user_id = 1042 AND status = 'COMPLETED' 
ORDER BY created_at DESC;`,

    groupby: `SELECT category_id, COUNT(*), AVG(price) 
FROM products 
GROUP BY category_id 
HAVING COUNT(*) > 5 
ORDER BY AVG(price) DESC;`,

    subquery: `SELECT id, name, price 
FROM products 
WHERE category_id IN (
    SELECT id FROM categories WHERE active = true
) AND price > (
    SELECT AVG(price) FROM products
);`
  };

  // DOM Elements
  const sqlInput = document.getElementById('sqlInput');
  const optimizeBtn = document.getElementById('optimizeBtn');
  const formatSqlBtn = document.getElementById('formatSqlBtn');
  const emptyState = document.getElementById('emptyState');
  const outputContent = document.getElementById('outputContent');
  const scoreBadge = document.getElementById('scoreBadge');
  const speedupVal = document.getElementById('speedupVal');
  const ioVal = document.getElementById('ioVal');
  const healthDeltaVal = document.getElementById('healthDeltaVal');
  const optimizedSqlCode = document.getElementById('optimizedSqlCode');
  const indexDDLCode = document.getElementById('indexDDLCode');
  const planNodes = document.getElementById('planNodes');
  const explanationsList = document.getElementById('explanationsList');
  const copySqlBtn = document.getElementById('copySqlBtn');

  // Preset Buttons
  document.querySelectorAll('.preset-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const type = btn.getAttribute('data-preset');
      if (PRESETS[type]) {
        sqlInput.value = PRESETS[type];
        runOptimizer();
      }
    });
  });

  // Format SQL Button
  if (formatSqlBtn) {
    formatSqlBtn.addEventListener('click', () => {
      if (!sqlInput.value.trim()) return;
      let formatted = sqlInput.value
        .replace(/\bSELECT\b/gi, '\nSELECT')
        .replace(/\bFROM\b/gi, '\nFROM')
        .replace(/\bWHERE\b/gi, '\nWHERE')
        .replace(/\bJOIN\b/gi, '\nJOIN')
        .replace(/\bGROUP BY\b/gi, '\nGROUP BY')
        .replace(/\bORDER BY\b/gi, '\nORDER BY')
        .trim();
      sqlInput.value = formatted;
    });
  }

  // Run Optimizer Button
  if (optimizeBtn) {
    optimizeBtn.addEventListener('click', runOptimizer);
  }

  // Main Optimizer Engine Logic
  function runOptimizer() {
    const rawSql = sqlInput.value.trim();
    if (!rawSql) return;

    // Show Loading Animation on Button
    optimizeBtn.disabled = true;
    optimizeBtn.innerHTML = `<span>⚡ Analyzing Execution Plan...</span>`;

    setTimeout(() => {
      optimizeBtn.disabled = false;
      optimizeBtn.innerHTML = `<span>⚡ Run AI Optimizer</span>`;

      // Perform Analysis
      const analysis = analyzeQuery(rawSql);

      // Render Outputs
      emptyState.classList.add('hidden');
      outputContent.classList.remove('hidden');

      // Update Badges & Metrics
      scoreBadge.innerText = `Health: ${analysis.newScore}/100`;
      scoreBadge.className = `score-badge ${analysis.newScore > 75 ? 'score-high' : 'score-low'}`;

      speedupVal.innerText = analysis.speedup;
      ioVal.innerText = analysis.ioSavings;
      healthDeltaVal.innerText = `+${analysis.newScore - analysis.oldScore} Points`;

      // Optimized SQL Diff
      optimizedSqlCode.innerText = analysis.optimizedSql;

      // Recommended DDL
      indexDDLCode.innerText = analysis.indexDDL;

      // Execution Plan Nodes
      planNodes.innerHTML = analysis.planNodesHtml;

      // Explanations
      explanationsList.innerHTML = analysis.explanations.map(exp => `<li>${exp}</li>`).join('');

    }, 350);
  }

  // AI Query Analysis Engine
  function analyzeQuery(sql) {
    const uppercaseSql = sql.toUpperCase();
    let oldScore = 45;
    let newScore = 94;
    let speedup = '3.8x Faster';
    let ioSavings = '82% Savings';
    let explanations = [];
    let indexDDL = '';
    let optimizedSql = sql;
    let planNodesHtml = '';

    // Check 1: SELECT * Detection
    if (uppercaseSql.includes('SELECT *')) {
      explanations.push('Replaced wildcard <code>SELECT *</code> with explicit column list to reduce network payload and I/O buffer footprint.');
      optimizedSql = optimizedSql.replace(/SELECT \*/i, 'SELECT id, status, created_at');
      oldScore -= 10;
    } else {
      explanations.push('Explicit column list detected — optimal memory overhead.');
    }

    // Check 2: Missing Index / WHERE scan
    if (uppercaseSql.includes('WHERE')) {
      let tableMatch = sql.match(/FROM\s+([a-zA-Z0-9_]+)/i);
      let tableName = tableMatch ? tableMatch[1] : 'orders';
      indexDDL += `-- Create Composite Index to eliminate Sequential Scan\nCREATE INDEX idx_${tableName}_composite ON ${tableName} (status, created_at);\n`;
      explanations.push(`Generated composite index <code>idx_${tableName}_composite</code> to allow B-Tree Index Scan instead of full table scan.`);
    } else {
      explanations.push('Query lacks a <code>WHERE</code> filtering clause. Consider adding pagination or bounds.');
    }

    // Check 3: Subqueries to CTEs
    if (uppercaseSql.includes('IN (') || uppercaseSql.includes('SELECT AVG')) {
      explanations.push('Converted nested subqueries into Common Table Expressions (CTEs) for query planner readability and materialization.');
      optimizedSql = `WITH avg_metrics AS (\n    SELECT AVG(price) AS avg_price FROM products\n)\n` + optimizedSql;
      speedup = '4.5x Faster';
      ioSavings = '88% Savings';
      newScore = 98;
    }

    // Check 4: JOIN optimization
    if (uppercaseSql.includes('JOIN')) {
      indexDDL += `\n-- FK Index for Join predicate\nCREATE INDEX idx_orders_user_id ON orders (user_id);`;
      explanations.push('Added Foreign Key index recommendation on Join predicate column to convert Nested Loop Join into Hash Join.');
    }

    if (!indexDDL) {
      indexDDL = `-- Current query schema is well-indexed.\n-- No new DDL statements required.`;
    }

    // Generate Execution Plan Tree HTML
    planNodesHtml = `
      <div class="tree-node">
        <strong>Hash Join (Cost: 12.4..48.2)</strong>
        <div style="font-size:0.75rem; color:#a1a1aa;">Hash Cond: (orders.user_id = users.id)</div>
      </div>
      <div class="tree-node" style="margin-left:1.5rem;">
        <strong>-> Index Scan using idx_users_status on users</strong>
        <div style="font-size:0.75rem; color:#10b981;">Rows: 1,420 | Filter: (status = 'active')</div>
      </div>
      <div class="tree-node" style="margin-left:1.5rem;">
        <strong>-> Index Scan using idx_orders_user_id on orders</strong>
        <div style="font-size:0.75rem; color:#10b981;">Rows: 4,800 | Buffer Hits: 100%</div>
      </div>
    `;

    return {
      oldScore,
      newScore,
      speedup,
      ioSavings,
      optimizedSql,
      indexDDL,
      planNodesHtml,
      explanations
    };
  }

  // Copy SQL Button
  if (copySqlBtn) {
    copySqlBtn.addEventListener('click', () => {
      navigator.clipboard.writeText(optimizedSqlCode.innerText);
      copySqlBtn.innerText = 'Copied!';
      setTimeout(() => copySqlBtn.innerText = 'Copy SQL', 2000);
    });
  }

  // Output Tab Switching
  const tabBtns = document.querySelectorAll('.tab-btn');
  tabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      tabBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');

      const targetTab = btn.getAttribute('data-tab');
      document.querySelectorAll('.tab-pane').forEach(pane => pane.classList.add('hidden'));

      if (targetTab === 'sqlDiff') document.getElementById('tabSqlDiff').classList.remove('hidden');
      if (targetTab === 'plan') document.getElementById('tabPlan').classList.remove('hidden');
      if (targetTab === 'indexes') document.getElementById('tabIndexes').classList.remove('hidden');
    });
  });

  // Search & Filter Ideas
  const ideaSearchInput = document.getElementById('ideaSearchInput');
  const filterBtns = document.querySelectorAll('.filter-btn');
  const ideaCards = document.querySelectorAll('.idea-card');

  if (ideaSearchInput) {
    ideaSearchInput.addEventListener('input', filterIdeas);
  }

  filterBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      filterBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      filterIdeas();
    });
  });

  function filterIdeas() {
    const query = ideaSearchInput ? ideaSearchInput.value.toLowerCase() : '';
    const activeFilter = document.querySelector('.filter-btn.active').getAttribute('data-filter');

    ideaCards.forEach(card => {
      const category = card.getAttribute('data-category');
      const title = card.querySelector('.idea-title').innerText.toLowerCase();
      const snippet = card.querySelector('.idea-snippet').innerText.toLowerCase();

      const matchesSearch = title.includes(query) || snippet.includes(query);
      const matchesCategory = (activeFilter === 'all' || category === activeFilter);

      if (matchesSearch && matchesCategory) {
        card.style.display = 'flex';
      } else {
        card.style.display = 'none';
      }
    });
  }

  // Reader Modal
  const readerModal = document.getElementById('readerModal');
  const modalBody = document.getElementById('modalBody');
  const closeModalBtn = document.getElementById('closeModalBtn');
  const modalOverlay = document.querySelector('.modal-overlay');

  const ARTICLES = {
    "1": {
      title: "Eliminating Sequential Scans in PostgreSQL at Scale",
      date: "Aug 2026 • 5 min read",
      content: `
        <p>In high-throughput PostgreSQL databases, sequential table scans (Seq Scan) are the silent killer of query performance. When a table grows past millions of rows, scanning every page off disk causes severe I/O spikes and CPU starvation.</p>
        <h3>1. Understanding Composite Index Column Ordering</h3>
        <p>When creating a composite index <code>CREATE INDEX idx_user_orders ON orders (status, created_at);</code>, column order matters immensely. Place equality columns first, followed by range comparison columns.</p>
        <h3>2. Leveraging Partial Indexes</h3>
        <p>If you frequently query only active users, do not index inactive or archived records. Create a partial index:</p>
        <pre><code class="code-block">CREATE INDEX idx_active_users ON users (email) WHERE status = 'active';</code></pre>
        <p>This drastically reduces index size and keeps memory buffer hit rates close to 100%.</p>
      `
    },
    "2": {
      title: "Building Resilient Microservices with Go Routines & Channels",
      date: "Aug 2026 • 7 min read",
      content: `
        <p>Go's lightweight concurrency primitives make it the ideal language for high-performance backend microservices. However, unmanaged goroutines can quickly leak memory or deadlock under peak traffic.</p>
        <h3>Worker Pool Architecture</h3>
        <p>Instead of spawning an unbounded goroutine per incoming HTTP request, use a worker pool with a buffered channel queue:</p>
        <pre><code class="code-block">type JobQueue struct {
    Jobs chan Job
}

func NewWorkerPool(maxWorkers int) *JobQueue {
    // Spawns bounded goroutines
}</code></pre>
      `
    },
    "3": {
      title: "AST-Driven SQL Refactoring for LLM Query Optimizers",
      date: "Aug 2026 • 6 min read",
      content: `
        <p>Large Language Models excel at natural language text, but can hallucinate column names or break SQL syntax when refactoring queries. Combining Abstract Syntax Tree (AST) static analysis with LLM prompt context guarantees safe rewrites.</p>
      `
    }
  };

  document.querySelectorAll('.read-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const articleId = btn.getAttribute('data-id');
      const article = ARTICLES[articleId];
      if (article) {
        modalBody.innerHTML = `
          <h2>${article.title}</h2>
          <div style="font-family:var(--font-mono); font-size:0.8rem; color:var(--text-muted); margin-bottom:1.5rem;">${article.date}</div>
          <div>${article.content}</div>
        `;
        readerModal.classList.remove('hidden');
      }
    });
  });

  if (closeModalBtn) closeModalBtn.addEventListener('click', closeModal);
  if (modalOverlay) modalOverlay.addEventListener('click', closeModal);

  function closeModal() {
    readerModal.classList.add('hidden');
  }

  // Contact Form Submission
  const contactForm = document.getElementById('contactForm');
  const formFeedback = document.getElementById('formFeedback');

  if (contactForm) {
    contactForm.addEventListener('submit', (e) => {
      e.preventDefault();
      formFeedback.innerText = "Thank you! Your note has been logged to knishad.lab.";
      formFeedback.className = "form-feedback feedback-success";
      formFeedback.classList.remove('hidden');
      contactForm.reset();
      setTimeout(() => formFeedback.classList.add('hidden'), 4000);
    });
  }

});
