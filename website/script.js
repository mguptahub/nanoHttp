// Theme handling
const themeToggle = document.querySelector('.theme-toggle');
const mobileThemeToggle = document.querySelector('.mobile-theme-toggle');
const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');

// Get current theme from localStorage or system preference
let currentTheme = localStorage.getItem('theme') || 
    (prefersDarkScheme.matches ? 'dark' : 'light');

// Set initial theme
document.documentElement.setAttribute('data-theme', currentTheme);
updateThemeIcons(currentTheme);

// Theme toggle click handlers
themeToggle.addEventListener('click', toggleTheme);
mobileThemeToggle.addEventListener('click', toggleTheme);

function toggleTheme() {
    currentTheme = currentTheme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', currentTheme);
    localStorage.setItem('theme', currentTheme);
    updateThemeIcons(currentTheme);
}

// Update theme icons based on current theme
function updateThemeIcons(theme) {
    const iconContent = theme === 'light' ? '🌙' : '☀️';
    themeToggle.innerHTML = iconContent;
    mobileThemeToggle.innerHTML = iconContent;
}

// Listen for system theme changes
prefersDarkScheme.addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
        currentTheme = e.matches ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', currentTheme);
        updateThemeIcons(currentTheme);
    }
});

// Scroll to Top functionality
const scrollToTopBtn = document.getElementById('scrollToTop');
const scrollThreshold = 300; // Show button after scrolling 300px

// Show/hide scroll to top button based on scroll position
window.addEventListener('scroll', () => {
    if (window.scrollY > scrollThreshold) {
        scrollToTopBtn.classList.add('show');
    } else {
        scrollToTopBtn.classList.remove('show');
    }
});

// Smooth scroll to top when button is clicked
scrollToTopBtn.addEventListener('click', () => {
    window.scrollTo({
        top: 0,
        behavior: 'smooth'
    });
});

// Copy functionality
document.querySelectorAll('.code-block').forEach(block => {
    const copyIcon = block.querySelector('.code-block-footer i');
    
    block.addEventListener('click', () => {
        const code = block.querySelector('code').textContent;
        navigator.clipboard.writeText(code).then(() => {
            // Change icon to check mark
            copyIcon.classList.remove('fa-copy');
            copyIcon.classList.add('fa-check');
            copyIcon.style.color = 'var(--terminal-prompt)';
            
            // Reset icon after 3 seconds
            setTimeout(() => {
                copyIcon.classList.remove('fa-check');
                copyIcon.classList.add('fa-copy');
                copyIcon.style.color = '';
            }, 3000);
        });
    });
});

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
    const promptSpan = line.querySelector('.prompt');
    const promptText = promptSpan.textContent;
    const originalText = line.textContent;
    const textWithoutPrompt = originalText.replace(promptText, '');
    
    // Save the prompt element
    const promptElement = promptSpan.cloneNode(true);
    
    // Clear line content but keep structure
    line.innerHTML = '';
    line.appendChild(promptElement);
    
    let charIndex = 0;

    function typeChar() {
        if (charIndex < textWithoutPrompt.length) {
            line.appendChild(document.createTextNode(textWithoutPrompt.charAt(charIndex)));
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