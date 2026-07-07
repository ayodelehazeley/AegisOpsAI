function App() {

return (

<div className="min-h-screen bg-gray-900 text-white p-10">

<h1 className="text-4xl font-bold">
AegisOps AI
</h1>

<p className="mt-4 text-xl">
Autonomous IT Operations Center
</p>


<div className="grid grid-cols-4 gap-6 mt-10">


<div className="bg-gray-800 p-6 rounded-xl">
<h2>Active Incidents</h2>
<p className="text-3xl">12</p>
</div>


<div className="bg-gray-800 p-6 rounded-xl">
<h2>Resolved Today</h2>
<p className="text-3xl">145</p>
</div>


<div className="bg-gray-800 p-6 rounded-xl">
<h2>AI Automation</h2>
<p className="text-3xl">
87%
</p>
</div>


<div className="bg-gray-800 p-6 rounded-xl">
<h2>System Health</h2>
<p className="text-green-400 text-3xl">
Healthy
</p>
</div>


</div>


</div>

)

}


export default App