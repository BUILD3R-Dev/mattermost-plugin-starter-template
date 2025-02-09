import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { Ticket, createTicket, updateTicket, listTickets, addComment } from '../api/ticketing';

type SortOrder = 'newest' | 'oldest';

const TicketingSystem: React.FC = () => {
    const [subject, setSubject] = useState('');
    const [description, setDescription] = useState('');
    const [tickets, setTickets] = useState<Ticket[]>([]);
    const [error, setError] = useState<string | null>(null);
    const [editingTicketId, setEditingTicketId] = useState<string | null>(null);
    const [editSubject, setEditSubject] = useState('');
    const [editDescription, setEditDescription] = useState('');
    const [newComment, setNewComment] = useState('');

    // New states for filtering, sorting and detailed view
    const [filterQuery, setFilterQuery] = useState('');
    const [sortOrder, setSortOrder] = useState<SortOrder>('newest');
    const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null);

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setError(null);
        try {
            await createTicket({ subject, description });
            setSubject('');
            setDescription('');
            fetchTickets();
        } catch (err: any) {
            setError(err.message);
        }
    };

    const handleUpdate = async (ticketId: string) => {
        setError(null);
        try {
            await updateTicket(ticketId, { subject: editSubject, description: editDescription });
            setEditingTicketId(null);
            fetchTickets();
        } catch (err: any) {
            setError(err.message);
        }
    };

    const handleAddComment = async (ticketId: string) => {
        setError(null);
        try {
            await addComment(ticketId, newComment);
            setNewComment('');
            fetchTickets();
        } catch (err: any) {
            setError(err.message);
        }
    };

    const fetchTickets = async () => {
        setError(null);
        try {
            const fetchedTickets = await listTickets();
            setTickets(fetchedTickets);
        } catch (err: any) {
            setError(err.message);
        }
    };

    // Filtering and sorting of tickets
    const filteredTickets = tickets.filter(ticket => 
        ticket.subject.toLowerCase().includes(filterQuery.toLowerCase())
    ).sort((a, b) => {
        const timeA = a.createdAt ?? 0;
        const timeB = b.createdAt ?? 0;
        return sortOrder === 'newest' ? timeB - timeA : timeA - timeB;
    });

    // Detailed ticket view: When a ticket is selected, show detailed view below list.
    const handleSelectTicket = (ticket: Ticket) => {
        setSelectedTicket(ticket);
    };

    useEffect(() => {
        fetchTickets();
    }, []);

    return (
        <div>
            <h2>Ticketing System</h2>
            {error && <div style={{ color: 'red' }}>Error: {error}</div>}

            <section>
                <h3>Create Ticket</h3>
                <form onSubmit={handleSubmit}>
                    <div>
                        <label htmlFor="subject">Subject:</label>
                        <input
                            id="subject"
                            name="subject"
                            type="text"
                            required
                            value={subject}
                            onChange={(e: ChangeEvent<HTMLInputElement>) => setSubject(e.target.value)}
                        />
                    </div>
                    <div>
                        <label htmlFor="description">Description:</label>
                        <textarea
                            id="description"
                            name="description"
                            required
                            value={description}
                            onChange={(e: ChangeEvent<HTMLTextAreaElement>) => setDescription(e.target.value)}
                        />
                    </div>
                    <button type="submit">Create Ticket</button>
                </form>
            </section>

            <section>
                <h3>Filter & Sort Tickets</h3>
                <div>
                    <input
                        type="text"
                        placeholder="Filter by subject..."
                        value={filterQuery}
                        onChange={(e: ChangeEvent<HTMLInputElement>) => setFilterQuery(e.target.value)}
                    />
                    <select value={sortOrder} onChange={(e: ChangeEvent<HTMLSelectElement>) => setSortOrder(e.target.value as SortOrder)}>
                        <option value="newest">Newest First</option>
                        <option value="oldest">Oldest First</option>
                    </select>
                </div>
            </section>

            <section>
                <h3>Ticket List</h3>
                <div>
                    {filteredTickets.length === 0 ? (
                        <p>No tickets available.</p>
                    ) : (
                        <ul>
                            {filteredTickets.map((ticket) => (
                                <li key={ticket.id} style={{ cursor: 'pointer' }} onClick={() => handleSelectTicket(ticket)}>
                                    <div>
                                        <strong>{ticket.subject}</strong> - {new Date(((ticket.createdAt ?? 0) * 1000)).toLocaleString()}
                                        <button onClick={(e) => { 
                                            e.stopPropagation(); 
                                            setEditingTicketId(ticket.id!);
                                            setEditSubject(ticket.subject);
                                            setEditDescription(ticket.description);
                                        }}>Edit</button>
                                    </div>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </section>

            {editingTicketId && (
                <section>
                    <h3>Edit Ticket</h3>
                    <div>
                        <input
                            type="text"
                            value={editSubject}
                            onChange={(e) => setEditSubject(e.target.value)}
                        />
                        <textarea
                            value={editDescription}
                            onChange={(e) => setEditDescription(e.target.value)}
                        />
                        <button onClick={() => handleUpdate(editingTicketId)}>Save</button>
                        <button onClick={() => setEditingTicketId(null)}>Cancel</button>
                    </div>
                </section>
            )}

            {selectedTicket && (
                <section>
                    <h3>Ticket Details</h3>
                    <div>
                        <p><strong>Subject:</strong> {selectedTicket.subject}</p>
                        <p><strong>Description:</strong> {selectedTicket.description}</p>
                        {selectedTicket.Comments && selectedTicket.Comments.length > 0 && (
                            <div>
                                <strong>Comments:</strong>
                                <ul>
                                    {selectedTicket.Comments.map((comment: string, index: number) => (
                                        <li key={index}>{comment}</li>
                                    ))}
                                </ul>
                            </div>
                        )}
                        <div>
                            <input
                                type="text"
                                placeholder="Add a comment..."
                                value={newComment}
                                onChange={(e: ChangeEvent<HTMLInputElement>) => setNewComment(e.target.value)}
                            />
                            <button onClick={() => selectedTicket.id && handleAddComment(selectedTicket.id)}>Add Comment</button>
                        </div>
                        <button onClick={() => setSelectedTicket(null)}>Close Details</button>
                    </div>
                </section>
            )}
        </div>
    );
};

export default TicketingSystem;
