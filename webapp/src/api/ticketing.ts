export interface Ticket {
    id?: string;
    subject: string;
    description: string;
    createdAt?: number;
    Comments?: string[];
}

export async function createTicket(ticket: Ticket): Promise<any> {
    const response = await fetch('/plugins/com.mattermost.sample.ticketing/api/v1/tickets', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(ticket)
    });
    if (!response.ok) {
        throw new Error(`Error creating ticket: ${response.statusText}`);
    }
    return response.json();
}

export async function updateTicket(ticketId: string, updatedData: Partial<Ticket>): Promise<any> {
    const response = await fetch(`/plugins/com.mattermost.sample.ticketing/api/v1/tickets/${ticketId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(updatedData)
    });
    if (!response.ok) {
        throw new Error(`Error updating ticket: ${response.statusText}`);
    }
    return response.json();
}

export async function listTickets(): Promise<Ticket[]> {
    const response = await fetch('/plugins/com.mattermost.sample.ticketing/api/v1/tickets', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    });
    if (!response.ok) {
        throw new Error(`Error fetching tickets: ${response.statusText}`);
    }
    return response.json();
}

export async function addComment(ticketId: string, comment: string): Promise<any> {
    const response = await fetch(`/plugins/com.mattermost.sample.ticketing/api/v1/tickets/${ticketId}/comment`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ comment })
    });
    if (!response.ok) {
        throw new Error(`Error adding comment: ${response.statusText}`);
    }
    return response.json();
}
