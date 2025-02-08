export const OPEN_TICKET_MODAL = 'OPEN_TICKET_MODAL';

export function openTicketModal() {
    return {
        type: OPEN_TICKET_MODAL,
        open: true
    };
}

export function closeTicketModal() {
    return {
        type: OPEN_TICKET_MODAL,
        open: false
    };
}
