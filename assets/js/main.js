/* ═══════════════════════════════════════
   Summer Party 2026 — Frontend JS
   ═══════════════════════════════════════ */

/* — Spell page: prefill from URL ——————— */
(function prefillSpell() {
  const input = document.getElementById('spell-input');
  if (!input) return;
  const params = new URLSearchParams(window.location.search);
  const spell = params.get('spell');
  if (spell) {
    input.value = spell.toUpperCase();
    // Auto-submit if spell param present
    // (optional: comment out to require manual submit)
    // document.getElementById('spell-form').submit();
  }
  input.addEventListener('input', function() {
    this.value = this.value.toUpperCase();
  });
})();

/* — RSVP form interactivity ——————————— */
(function rsvpForm() {
  const form = document.getElementById('rsvp-form');
  if (!form) return;

  // Plus one toggle
  const plusOneRadios = form.querySelectorAll('input[name="plus_one"]');
  const plusOneNameField = document.getElementById('plus-one-name-field');
  function updatePlusOne() {
    const val = form.querySelector('input[name="plus_one"]:checked')?.value;
    if (plusOneNameField) {
      plusOneNameField.classList.toggle('hidden', val !== 'yes');
    }
  }
  plusOneRadios.forEach(r => r.addEventListener('change', updatePlusOne));
  updatePlusOne();

  // Attending toggle — hide fields when declining
  const attendingRadios = form.querySelectorAll('input[name="attending"]');
  const optionalFields = document.getElementById('optional-fields');
  function updateAttending() {
    const val = form.querySelector('input[name="attending"]:checked')?.value;
    if (optionalFields) {
      optionalFields.classList.toggle('hidden', val === 'no');
    }
  }
  attendingRadios.forEach(r => r.addEventListener('change', updateAttending));
  updateAttending();

  // Newsletter toggle — show only when email has value
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

/* — Scroll reveal ——————————————————————— */
(function scrollReveal() {
  const revealEls = document.querySelectorAll('.content-card, .rsvp-section');
  if (!revealEls.length) return;

  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry, i) => {
          if (entry.isIntersecting) {
            setTimeout(() => {
              entry.target.classList.add('revealed');
            }, i * 80);
            observer.unobserve(entry.target);
          }
        });
      },
      { threshold: 0.1, rootMargin: '0px 0px -40px 0px' }
    );
    revealEls.forEach(el => observer.observe(el));
  } else {
    // Fallback: reveal all immediately
    revealEls.forEach(el => el.classList.add('revealed'));
  }
})();

/* — Admin: confirm destructive actions — */
(function adminConfirm() {
  document.querySelectorAll('[data-confirm]').forEach(el => {
    el.addEventListener('click', function(e) {
      if (!window.confirm(this.dataset.confirm)) {
        e.preventDefault();
      }
    });
  });
})();

/* — Admin: copy to clipboard ————————— */
(function copyButtons() {
  document.querySelectorAll('[data-copy]').forEach(btn => {
    btn.addEventListener('click', function() {
      const text = this.dataset.copy;
      navigator.clipboard.writeText(text).then(() => {
        const orig = this.textContent;
        this.textContent = '✓';
        setTimeout(() => { this.textContent = orig; }, 1500);
      });
    });
  });
})();
