/**
 * knishad.labs — Interactive Client Engine
 * Theme Switching, R&D Field Notes & Dynamic UI Controls
 */

document.addEventListener('DOMContentLoaded', () => {

  // ==========================================
  // 1. Theme Toggle Engine (Light / Dark Mode)
  // ==========================================
  const themeToggleBtn = document.getElementById('themeToggleBtn');
  const themeIcon = document.getElementById('themeIcon');
  const themeText = document.getElementById('themeText');
  const htmlEl = document.documentElement;

  // Check saved theme preference or default to light
  const savedTheme = localStorage.getItem('theme') || 'light';
  applyTheme(savedTheme);

  if (themeToggleBtn) {
    themeToggleBtn.addEventListener('click', () => {
      const currentTheme = htmlEl.classList.contains('dark') ? 'dark' : 'light';
      const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
      applyTheme(newTheme);
      localStorage.setItem('theme', newTheme);
    });
  }

  function applyTheme(theme) {
    if (theme === 'dark') {
      htmlEl.classList.remove('light');
      htmlEl.classList.add('dark');
      if (themeIcon) themeIcon.innerText = '🌙';
      if (themeText) themeText.innerText = 'Dark';
    } else {
      htmlEl.classList.remove('dark');
      htmlEl.classList.add('light');
      if (themeIcon) themeIcon.innerText = '☀️';
      if (themeText) themeText.innerText = 'Light';
    }
  }

  // ==========================================
  // 2. Search & Filter Technical Notes
  // ==========================================
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

  // ==========================================
  // 3. Reader Modal Popup
  // ==========================================
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
      title: "Deploying Crunchy Postgres on Kubernetes Clusters",
      date: "Aug 2026 • 7 min read",
      content: `
        <p>Managing enterprise production PostgreSQL on Kubernetes requires robust operators. Crunchy Postgres (PGO) provides declarative management for high availability, backup automation, and disaster recovery.</p>
        <h3>1. StatefulSet Storage Architecture</h3>
        <p>Configuring PersistentVolumeClaims (PVCs) with high-IOPS storage classes guarantees that database write speed remains unthrottled under heavy write loads.</p>
        <h3>2. Automated Failover with pgBackRest</h3>
        <p>Crunchy PGO integrates pgBackRest natively, enabling point-in-time recovery (PITR) and automatic failover without manual intervention.</p>
      `
    },
    "3": {
      title: "Building Resilient Microservices with Go Routines & Channels",
      date: "Aug 2026 • 6 min read",
      content: `
        <p>Go's lightweight concurrency primitives make it the ideal language for high-performance backend microservices. However, unmanaged goroutines can quickly leak memory or deadlock under peak traffic.</p>
        <h3>Worker Pool Architecture</h3>
        <p>Instead of spawning an unbounded goroutine per incoming HTTP request, use a worker pool with a buffered channel queue to process tasks safely.</p>
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

  // ==========================================
  // 4. Contact Form Submission
  // ==========================================
  const contactForm = document.getElementById('contactForm');
  const formFeedback = document.getElementById('formFeedback');

  if (contactForm) {
    contactForm.addEventListener('submit', (e) => {
      e.preventDefault();
      formFeedback.innerText = "Thank you! Your message has been sent to knishad.labs.";
      formFeedback.className = "form-feedback feedback-success";
      formFeedback.classList.remove('hidden');
      contactForm.reset();
      setTimeout(() => formFeedback.classList.add('hidden'), 4000);
    });
  }

});
