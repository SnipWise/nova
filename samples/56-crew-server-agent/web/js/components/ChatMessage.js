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
            let classes = `message ${props.message.role}`;
            // Ajouter la classe 'processing' pour les messages information non complétés
            if (props.message.role === 'information' && !props.message.completed) {
                classes += ' processing';
            }
            return classes;
        });

        const roleDisplay = Vue.computed(() => {
            const role = props.message.role || 'unknown';
            return role.charAt(0).toUpperCase() + role.slice(1);
        });

        const showLoader = Vue.computed(() => {
            // Afficher le loader pour assistant en streaming OU information en cours
            return (props.isStreaming && props.message.role === 'assistant') ||
                   (props.message.role === 'information' && !props.message.completed);
        });

        return {
            renderContent,
            messageClass,
            roleDisplay,
            showLoader
        };
    },

    template: `
        <div :class="messageClass">
            <div class="message-role">
                {{ roleDisplay }}
                <span v-if="showLoader" class="role-loader"></span>
            </div>
            <div
                class="message-content"
                v-html="renderContent"
            ></div>
        </div>
    `
};

// Register component globally
window.ChatMessage = ChatMessage;
