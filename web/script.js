class SalesTracker {
    constructor() {
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadSales();
    }

    bindEvents() {
        document.getElementById('saleForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addSale();
        });

        document.getElementById('analyticsForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.getAnalytics();
        });

        document.getElementById('search').addEventListener('input', (e) => {
            this.filterSales(e.target.value);
        });
    }

    async addSale() {
        const sale = {
            type: document.getElementById('type').value,
            amount: parseFloat(document.getElementById('amount').value),
            date: new Date(document.getElementById('date').value).toISOString(),
            category: document.getElementById('category').value
        };

        try {
            const response = await fetch('/api/items', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(sale)
            });

            if (response.ok) {
                this.resetForm();
                this.loadSales();
            } else {
                alert('Error adding sale');
            }
        } catch (error) {
            alert('Error adding sale');
        }
    }

    async loadSales() {
        try {
            const response = await fetch('/api/items');
            const sales = await response.json();
            this.renderSales(sales);
        } catch (error) {
            alert('Error loading sales');
        }
    }

    renderSales(sales) {
        const tbody = document.querySelector('#salesTable tbody');
        tbody.innerHTML = '';

        sales.forEach(sale => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${sale.id}</td>
                <td>${sale.type}</td>
                <td>${sale.amount}</td>
                <td>${new Date(sale.date).toLocaleString()}</td>
                <td>${sale.category}</td>
                <td>
                    <button onclick="salesTracker.editSale(${sale.id})">Edit</button>
                    <button onclick="salesTracker.deleteSale(${sale.id})">Delete</button>
                </td>
            `;
            tbody.appendChild(row);
        });
    }

    filterSales(query) {
        const rows = document.querySelectorAll('#salesTable tbody tr');
        rows.forEach(row => {
            const text = row.textContent.toLowerCase();
            row.style.display = text.includes(query.toLowerCase()) ? '' : 'none';
        });
    }

    async getAnalytics() {
        const from = document.getElementById('from').value;
        const to = document.getElementById('to').value;

        if (!from || !to) {
            alert('Please select both dates');
            return;
        }

        try {
            const fromDate = new Date(from).toISOString();
            const toDate = new Date(to).toISOString();
            const response = await fetch(`/api/analytics?from=${encodeURIComponent(fromDate)}&to=${encodeURIComponent(toDate)}`);
            const analytics = await response.json();
            this.renderAnalytics(analytics);
        } catch (error) {
            alert('Error loading analytics');
        }
    }

    renderAnalytics(analytics) {
        document.getElementById('totalSales').textContent = analytics.sum.toFixed(2);
        document.getElementById('average').textContent = analytics.average.toFixed(2);
        document.getElementById('count').textContent = analytics.count;
        document.getElementById('median').textContent = analytics.median.toFixed(2);
        document.getElementById('percentile90').textContent = analytics.percentile90.toFixed(2);
    }

    resetForm() {
        document.getElementById('saleForm').reset();
    }

    async editSale(id) {
        // Implementation for edit
        alert('Edit functionality not implemented yet');
    }

    async deleteSale(id) {
        if (confirm('Are you sure you want to delete this sale?')) {
            try {
                const response = await fetch(`/api/items/${id}`, {
                    method: 'DELETE'
                });

                if (response.ok) {
                    this.loadSales();
                } else {
                    alert('Error deleting sale');
                }
            } catch (error) {
                alert('Error deleting sale');
            }
        }
    }
}

const salesTracker = new SalesTracker();