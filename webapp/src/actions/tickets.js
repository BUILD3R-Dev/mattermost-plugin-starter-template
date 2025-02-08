export const RECEIVED_TICKETS = 'RECEIVED_TICKETS';

export function getTickets() {
    return async (dispatch) => {
        try {
            const response = await fetch('/plugins/com.mattermost.support-tickets/api/v1/tickets');
            const tickets = await response.json();
            dispatch({type: RECEIVED_TICKETS, data: tickets});
        } catch (error) {
            console.error('Failed to fetch tickets:', error); // eslint-disable-line no-console
        }
    };
}
