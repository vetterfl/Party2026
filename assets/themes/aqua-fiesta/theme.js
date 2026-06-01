/* ═══════════════════════════════════════════════
   AQUA FIESTA — Theme JS
   Extends shared behaviours with tropical extras
   ═══════════════════════════════════════════════ */

/* ── Spell page: prefill from URL ──────────────── */
(function prefillSpell() {
  const input = document.getElementById('spell-input');
  if (!input) return;
  const params = new URLSearchParams(window.location.search);
  const spell = params.get('spell');
  if (spell) input.value = spell.toUpperCase();
  input.addEventListener('input', function () {
    this.value = this.value.toUpperCase();
  });
})();

/* ── Spell page: bubble particles ──────────────── */
(function spellBubbles() {
  const page = document.querySelector('.spell-page');
  if (!page) return;

  const colors = ['rgba(255,255,255,0.25)', 'rgba(255,182,39,0.35)', 'rgba(255,92,56,0.25)', 'rgba(0,210,188,0.30)'];

  function spawnBubble() {
    const el = document.createElement('span');
    const size = 8 + Math.random() * 24;
    const startX = Math.random() * 100;
    const duration = 6 + Math.random() * 8;
    const delay = Math.random() * 2;
    const color = colors[Math.floor(Math.random() * colors.length)];

    Object.assign(el.style, {
      position: 'absolute',
      bottom: '-40px',
      left: startX + '%',
      width: size + 'px',
      height: size + 'px',
      borderRadius: '50%',
      background: color,
      border: '1px solid rgba(255,255,255,0.3)',
      pointerEvents: 'none',
      zIndex: '0',
      animation: `bubble-rise ${duration}s ${delay}s ease-in forwards`,
    });

    page.appendChild(el);
    setTimeout(() => el.remove(), (duration + delay + 0.5) * 1000);
  }

  // Inject keyframe if not already present
  if (!document.getElementById('aq-bubble-kf')) {
    const s = document.createElement('style');
    s.id = 'aq-bubble-kf';
    s.textContent = `
      @keyframes bubble-rise {
        0%   { transform: translateY(0) scale(1);       opacity: 0; }
        10%  { opacity: 1; }
        80%  { opacity: 0.8; }
        100% { transform: translateY(-110vh) scale(0.4); opacity: 0; }
      }
    `;
    document.head.appendChild(s);
  }

  // Spawn bubbles periodically
  setInterval(spawnBubble, 700);
  // Initial burst
  for (let i = 0; i < 6; i++) setTimeout(spawnBubble, i * 200);
})();

/* ── RSVP form interactivity ────────────────────── */
(function rsvpForm() {
  const form = document.getElementById('rsvp-form');
  if (!form) return;

  const plusOneRadios = form.querySelectorAll('input[name="plus_one"]');
  const plusOneNameField = document.getElementById('plus-one-name-field');
  function updatePlusOne() {
    const val = form.querySelector('input[name="plus_one"]:checked')?.value;
    if (plusOneNameField) plusOneNameField.classList.toggle('hidden', val !== 'yes');
  }
  plusOneRadios.forEach(r => r.addEventListener('change', updatePlusOne));
  updatePlusOne();

  const attendingRadios = form.querySelectorAll('input[name="attending"]');
  const optionalFields = document.getElementById('optional-fields');
  function updateAttending() {
    const val = form.querySelector('input[name="attending"]:checked')?.value;
    if (optionalFields) optionalFields.classList.toggle('hidden', val === 'no');
  }
  attendingRadios.forEach(r => r.addEventListener('change', updateAttending));
  updateAttending();

  const emailInput = document.getElementById('email-input');
  const newsletterField = document.getElementById('newsletter-field');
  function updateNewsletter() {
    if (!newsletterField || !emailInput) return;
    newsletterField.classList.toggle('hidden', emailInput.value.trim() === '');
  }
  if (emailInput) {
    emailInput.addEventListener('input', updateNewsletter);
    updateNewsletter();
  }
})();

/* ── Confetti on confirmed page ─────────────────── */
(function confirmConfetti() {
  if (!document.querySelector('.confirmed-page')) return;

  const colors = ['#FF5C38', '#FFB627', '#00C9B5', '#FF6B9D', '#fff'];

  function burst() {
    const count = 60;
    for (let i = 0; i < count; i++) {
      const el = document.createElement('div');
      el.className = 'confetti-piece';
      const size = 6 + Math.random() * 10;
      const x = 20 + Math.random() * 60; // % from left
      const duration = 1.8 + Math.random() * 1.5;
      const color = colors[Math.floor(Math.random() * colors.length)];
      const shapes = ['50%', '0%', '50% 0 0 50%'];
      Object.assign(el.style, {
        left: x + 'vw',
        top: '-10px',
        width: size + 'px',
        height: size * (0.5 + Math.random()) + 'px',
        background: color,
        borderRadius: shapes[Math.floor(Math.random() * shapes.length)],
        '--duration': duration + 's',
        animationDelay: Math.random() * 0.4 + 's',
      });
      document.body.appendChild(el);
      setTimeout(() => el.remove(), (duration + 0.8) * 1000);
    }
  }

  // Small delay so page has loaded
  setTimeout(burst, 600);
  setTimeout(burst, 1400);
})();

/* ── Scroll reveal ──────────────────────────────── */
(function scrollReveal() {
  const revealEls = document.querySelectorAll('.content-card, .rsvp-section');
  if (!revealEls.length) return;

  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry, i) => {
          if (entry.isIntersecting) {
            setTimeout(() => entry.target.classList.add('revealed'), i * 90);
            observer.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.08, rootMargin: '0px 0px -30px 0px' }
    );
    revealEls.forEach(el => observer.observe(el));
  } else {
    revealEls.forEach(el => el.classList.add('revealed'));
  }
})();

/* ── Admin: confirm destructive actions ─────────── */
(function adminConfirm() {
  document.querySelectorAll('[data-confirm]').forEach(el => {
    el.addEventListener('click', function (e) {
      if (!window.confirm(this.dataset.confirm)) e.preventDefault();
    });
  });
})();

/* ── Admin: copy to clipboard ────────────────────── */
(function copyButtons() {
  document.querySelectorAll('[data-copy]').forEach(btn => {
    btn.addEventListener('click', function () {
      navigator.clipboard.writeText(this.dataset.copy).then(() => {
        const orig = this.textContent;
        this.textContent = '✓';
        setTimeout(() => { this.textContent = orig; }, 1500);
      });
    });
  });
})();
