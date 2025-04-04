// Theme handling
const themeToggle = document.querySelector('.theme-toggle');
const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');

// Check for saved theme preference or use system preference
const currentTheme = localStorage.getItem('theme') || 
    (prefersDarkScheme.matches ? 'dark' : 'light');

// Set initial theme
document.documentElement.setAttribute('data-theme', currentTheme);
updateThemeIcon(currentTheme);

// Theme toggle click handler
themeToggle.addEventListener('click', () => {
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeIcon(newTheme);
});

// Update theme icon based on current theme
function updateThemeIcon(theme) {
    themeToggle.innerHTML = theme === 'light' ? '🌙' : '☀️';
}

// Copy button functionality
const copyBtn = document.querySelector('.copy-btn');
if (copyBtn) {
    copyBtn.addEventListener('click', () => {
        const code = document.querySelector('.install-command code').textContent;
        navigator.clipboard.writeText(code).then(() => {
            copyBtn.textContent = '✓';
            setTimeout(() => {
                copyBtn.textContent = '📋';
            }, 2000);
        });
    });
}

// Tab functionality
const tabButtons = document.querySelectorAll('.tab-btn');
const tabContents = document.querySelectorAll('.tab-content');

tabButtons.forEach(button => {
    button.addEventListener('click', () => {
        // Remove active class from all buttons and contents
        tabButtons.forEach(btn => btn.classList.remove('active'));
        tabContents.forEach(content => content.classList.remove('active'));

        // Add active class to clicked button and corresponding content
        button.classList.add('active');
        const tabId = button.getAttribute('data-tab');
        document.querySelector(`#${tabId}`).classList.add('active');
    });
});

// Smooth scrolling for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});

// Intersection Observer for fade-in animations
const observerOptions = {
    root: null,
    rootMargin: '0px',
    threshold: 0.1
};

const observer = new IntersectionObserver((entries, observer) => {
    entries.forEach(entry => {
        if (entry.isIntersecting) {
            entry.target.classList.add('fade-in');
            observer.unobserve(entry.target);
        }
    });
}, observerOptions);

// Observe elements with fade-in class
document.querySelectorAll('.feature-card, .install-option, .usage-example').forEach(element => {
    element.classList.add('fade-in-hidden');
    observer.observe(element);
});

// Terminal typing animation
const terminalLines = document.querySelectorAll('.terminal-line');
let currentLine = 0;

function typeLine(line) {
    const text = line.textContent;
    line.textContent = '';
    let charIndex = 0;

    function typeChar() {
        if (charIndex < text.length) {
            line.textContent += text.charAt(charIndex);
            charIndex++;
            setTimeout(typeChar, 50);
        } else {
            currentLine++;
            if (currentLine < terminalLines.length) {
                setTimeout(() => typeLine(terminalLines[currentLine]), 500);
            }
        }
    }

    typeChar();
}

// Start typing animation if terminal exists
if (terminalLines.length > 0) {
    setTimeout(() => typeLine(terminalLines[0]), 1000);
} 