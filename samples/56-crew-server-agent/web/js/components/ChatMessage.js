/**
 * ChatMessage Component
 * Renders individual chat messages with markdown support
 */

const ChatMessage = {
    name: 'ChatMessage',

    props: {
        message: {
            type: Object,
            required: true
        },
        isStreaming: {
            type: Boolean,
            default: false
        }
    },

    setup(props) {
        const markdownRenderer = new MarkdownRenderer();

        const renderContent = Vue.computed(() => {
            const content = props.message.content || '';

            // Use streaming renderer if message is being streamed
            if (props.isStreaming) {
                return markdownRenderer.renderStreaming(content);
            }

            return markdownRenderer.render(content);
        });

        const messageClass = Vue.computed(() => {
            return `message ${props.message.role}`;
        });

        const roleDisplay = Vue.computed(() => {
            const role = props.message.role || 'unknown';
            return role.charAt(0).toUpperCase() + role.slice(1);
        });

        return {
            renderContent,
            messageClass,
            roleDisplay
        };
    },

    template: `
        <div :class="messageClass">
            <div class="message-role">{{ roleDisplay }}</div>
            <div
                class="message-content"
                v-html="renderContent"
            ></div>
        </div>
    `
};

// Register component globally
window.ChatMessage = ChatMessage;
