/**
 * Markdown Rendering and Syntax Highlighting Utilities
 */

class MarkdownRenderer {
    constructor() {
        this.initializeMarked();
    }

    /**
     * Initialize marked.js with custom configuration
     */
    initializeMarked() {
        // Configure marked options
        marked.setOptions({
            breaks: true,
            gfm: true,
            headerIds: false,
            mangle: false,
            highlight: (code, lang) => {
                // Use highlight.js for syntax highlighting
                if (lang && hljs.getLanguage(lang)) {
                    try {
                        return hljs.highlight(code, { language: lang }).value;
                    } catch (e) {
                        console.error('Highlight error:', e);
                    }
                }
                // Auto-detect language if not specified
                try {
                    return hljs.highlightAuto(code).value;
                } catch (e) {
                    console.error('Auto-highlight error:', e);
                    return this.escapeHtml(code);
                }
            }
        });

        // Custom renderer for code blocks to ensure proper structure
        const renderer = new marked.Renderer();

        // Override code block rendering
        renderer.code = (code, language) => {
            const lang = language || 'plaintext';
            const validLang = hljs.getLanguage(lang) ? lang : 'plaintext';

            let highlightedCode;
            try {
                highlightedCode = hljs.highlight(code, { language: validLang }).value;
            } catch (e) {
                highlightedCode = this.escapeHtml(code);
            }

            // Store code in base64 to avoid escaping issues
            const base64Code = btoa(encodeURIComponent(code));

            return `<div class="code-block-wrapper"><button class="copy-code-btn" data-code-base64="${base64Code}" title="Copy code">Copy</button><pre><code class="hljs language-${this.escapeHtml(validLang)}">${highlightedCode}</code></pre></div>`;
        };

        // Override inline code rendering
        renderer.codespan = (code) => {
            return `<code>${this.escapeHtml(code)}</code>`;
        };

        // Set custom renderer
        marked.use({ renderer });
    }

    /**
     * Convert markdown text to HTML
     * @param {string} markdownText - Markdown text to convert
     * @returns {string} - HTML output
     */
    render(markdownText) {
        if (!markdownText) {
            return '';
        }

        try {
            return marked.parse(markdownText);
        } catch (e) {
            console.error('Markdown parsing error:', e);
            return this.escapeHtml(markdownText);
        }
    }

    /**
     * Render markdown incrementally (for streaming)
     * Handles incomplete code blocks gracefully
     * @param {string} markdownText - Markdown text to convert
     * @returns {string} - HTML output
     */
    renderStreaming(markdownText) {
        if (!markdownText) {
            return '';
        }

        try {
            // Check if there's an incomplete code block
            const codeBlockMatches = markdownText.match(/```/g);
            const hasIncompleteCodeBlock = codeBlockMatches && codeBlockMatches.length % 2 !== 0;

            if (hasIncompleteCodeBlock) {
                // Temporarily close the code block for rendering
                const textWithClosedBlock = markdownText + '\n```';
                let html = marked.parse(textWithClosedBlock);
                // Add a visual indicator for incomplete code
                html = html.replace(/<\/code><\/pre>$/, '<span class="cursor">▊</span></code></pre>');
                return html;
            }

            return marked.parse(markdownText);
        } catch (e) {
            console.error('Streaming markdown parsing error:', e);
            return this.escapeHtml(markdownText);
        }
    }

    /**
     * Escape HTML special characters
     * @param {string} text - Text to escape
     * @returns {string} - Escaped text
     */
    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;'
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }

    /**
     * Extract language from code block
     * @param {string} text - Text potentially containing code block
     * @returns {string|null} - Language or null
     */
    extractLanguage(text) {
        const match = text.match(/```(\w+)/);
        return match ? match[1] : null;
    }

    /**
     * Check if text contains code blocks
     * @param {string} text - Text to check
     * @returns {boolean}
     */
    hasCodeBlocks(text) {
        return /```/.test(text);
    }

    /**
     * Count number of code blocks
     * @param {string} text - Text to analyze
     * @returns {number}
     */
    countCodeBlocks(text) {
        const matches = text.match(/```/g);
        return matches ? Math.floor(matches.length / 2) : 0;
    }
}

// Export for use in other modules
window.MarkdownRenderer = MarkdownRenderer;

/**
 * Initialize copy button functionality
 * Attaches event listeners to copy code buttons
 */
document.addEventListener('click', (e) => {
    if (e.target.classList.contains('copy-code-btn')) {
        const button = e.target;
        const base64Code = button.getAttribute('data-code-base64');

        // Decode from base64
        const decodedCode = decodeURIComponent(atob(base64Code));

        // Copy to clipboard
        navigator.clipboard.writeText(decodedCode).then(() => {
            // Change button text and style
            const originalText = button.textContent;
            button.textContent = 'Copied!';
            button.classList.add('copied');

            // Reset after 2 seconds
            setTimeout(() => {
                button.textContent = originalText;
                button.classList.remove('copied');
            }, 2000);
        }).catch(err => {
            console.error('Failed to copy code:', err);
            button.textContent = 'Failed';
            setTimeout(() => {
                button.textContent = 'Copy';
            }, 2000);
        });
    }
});
