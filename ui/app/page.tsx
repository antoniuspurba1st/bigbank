const summaryCards = [
  { title: "Total Transactions", value: "1,234" },
  { title: "Total Volume", value: "$5,678,901.23" },
  { title: "Accounts", value: "3" },
];

const recentTransactions = [
  {
    id: "txn_1",
    reference: "ref-trf-001",
    amount: "$150.25",
    status: "Completed",
  },
  {
    id: "txn_2",
    reference: "ref-trf-002",
    amount: "$1,200.00",
    status: "Completed",
  },
  {
    id: "txn_3",
    reference: "ref-trf-003",
    amount: "$5,000,000",
    status: "Rejected",
  },
  {
    id: "txn_4",
    reference: "ref-trf-004",
    amount: "$75.50",
    status: "Completed",
  },
];

export default function Home() {
  return (
    <div>
      <h1 className="text-3xl font-semibold tracking-tight">Dashboard</h1>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
        {summaryCards.map((card) => (
          <div key={card.title} className="glass-panel p-6 rounded-lg">
            <h3 className="eyebrow">{card.title}</h3>
            <p className="text-2xl font-semibold mt-2">{card.value}</p>
          </div>
        ))}
      </div>
      <div className="mt-8">
        <h2 className="text-xl font-semibold tracking-tight">
          Recent Transactions
        </h2>
        <div className="mt-4 glass-panel rounded-lg p-6">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[color:var(--line)]">
                <th className="text-left py-2 eyebrow">Reference</th>
                <th className="text-left py-2 eyebrow">Amount</th>
                <th className="text-left py-2 eyebrow">Status</th>
              </tr>
            </thead>
            <tbody>
              {recentTransactions.map((txn) => (
                <tr key={txn.id} className="border-b border-[color:var(--line)]">
                  <td className="py-3 font-mono text-sm">{txn.reference}</td>
                  <td className="py-3">{txn.amount}</td>
                  <td className="py-3">
                    <span
                      className={`status-pill px-2 py-1 rounded-full text-xs ${
                        txn.status === "Completed"
                          ? "status-success"
                          : "status-rejected"
                      }`}
                    >
                      {txn.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
