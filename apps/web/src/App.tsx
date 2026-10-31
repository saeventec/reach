// The main component of the web app. Learn more: docs/learn/react.md
// The class names are Tailwind CSS. Learn more: docs/learn/tailwind.md
import { fetchNewestGreeting } from './greeting/greetingClient.ts'
import { NewestGreeting } from './greeting/NewestGreeting.tsx'

function App() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 bg-slate-50">
      <h1 className="text-4xl font-bold text-slate-900">Reach</h1>
      <NewestGreeting loadGreeting={fetchNewestGreeting} />
    </main>
  )
}

export default App
