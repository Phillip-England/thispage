package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// New creates a new directory with the given name, and standard subdirectories and files.
func New(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

    if force {
        if err := os.RemoveAll(name); err != nil {
             return fmt.Errorf("failed to remove existing directory '%s': %w", name, err)
        }
    }

	// Create the main project directory
	err := os.Mkdir(name, 0755)
	if err != nil {
        if os.IsExist(err) {
             return fmt.Errorf("directory '%s' already exists. Use --force to overwrite", name)
        }
		return fmt.Errorf("could not create project directory '%s': %w", name, err)
	}

	// Define subdirectory paths
	dirs := []string{"live", "partials", "templates", "templates/posts", "static", "layouts"}

	for _, dir := range dirs {
		dirPath := filepath.Join(name, dir)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create subdirectory '%s': %w", dir, err)
		}
	}

	templatesDirPath := filepath.Join(name, "templates")
	partialsDirPath := filepath.Join(name, "partials")
    layoutsDirPath := filepath.Join(name, "layouts")

    guestLayoutHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>{{ slot "title" }}</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen relative">
  <div class="max-w-4xl mx-auto py-20 px-6">
    {{ include "./partials/navigation.html" }}
    
    <main class="mt-12 min-h-[50vh] relative">
        {{ slot "main" }}
    </main>
  </div>
  
  <!-- The Add Content Button (Visible only in Admin Mode) -->
  <div class="thispage-add-btn hidden fixed bottom-8 right-8 z-50 cursor-pointer hover:scale-110 transition-transform group">
        <div class="w-14 h-14 rounded-full bg-blue-600 flex items-center justify-center text-white text-3xl font-bold shadow-lg hover:bg-blue-500 transition-colors">
            +
        </div>
        <div class="absolute right-full mr-4 top-1/2 -translate-y-1/2 bg-neutral-900 text-white text-xs px-3 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
            Add Block
        </div>
  </div>
</body>
</html>`

	defaultIndexHTML := `{{ layout "./layouts/guest_layout.html" }}

{{ block "title" }}Home - This Page{{ endblock }}

{{ block "main" }}
    {{ include "./partials/hero.html" }}
    
    <section class="grid grid-cols-1 md:grid-cols-3 gap-8 py-12">
        {{ include "./partials/feature_item.html" icon="⚡" title="Fast" description="Built for speed and performance from the ground up." }}
        {{ include "./partials/feature_item.html" icon="🎨" title="Beautiful" description="Stunning default styles using Tailwind CSS." }}
        {{ include "./partials/feature_item.html" icon="🛠" title="Flexible" description="Easy to customize and extend for your needs." }}
    </section>

    {{ include "./partials/content.html" }}
    {{ include "./partials/cta.html" title="Ready to start?" text="Get your project up and running in seconds." button_text="Get Started" button_link="/pricing" }}
    {{ include "./partials/footer.html" }}
{{ endblock }}

{{ endlayout }}`

    aboutPageHTML := `{{ layout "./layouts/guest_layout.html" }}
{{ block "title" }}About Us - This Page{{ endblock }}
{{ block "main" }}
    {{ include "./partials/page_header.html" title="About Our Team" subtitle="We are building the future of content management." }}
    
    <section class="grid grid-cols-1 md:grid-cols-3 gap-12 py-12">
        {{ include "./partials/team_member.html" name="Sarah Connor" role="Founder & CEO" image="https://i.pravatar.cc/300?img=1" }}
        {{ include "./partials/team_member.html" name="John Smith" role="CTO" image="https://i.pravatar.cc/300?img=3" }}
        {{ include "./partials/team_member.html" name="Emily Blunt" role="Head of Design" image="https://i.pravatar.cc/300?img=5" }}
    </section>
    
    {{ include "./partials/content.html" }}
    {{ include "./partials/footer.html" }}
{{ endblock }}
{{ endlayout }}`

    pricingPageHTML := `{{ layout "./layouts/guest_layout.html" }}
{{ block "title" }}Pricing - This Page{{ endblock }}
{{ block "main" }}
    {{ include "./partials/page_header.html" title="Simple Pricing" subtitle="Choose the plan that fits your needs." }}
    
    <section class="grid grid-cols-1 md:grid-cols-3 gap-8 py-12">
        {{ include "./partials/pricing_card.html" plan="Starter" price="0" description="Perfect for personal projects." }}
        {{ include "./partials/pricing_card.html" plan="Pro" price="29" description="For growing teams and businesses." }}
        {{ include "./partials/pricing_card.html" plan="Enterprise" price="99" description="Advanced features for large scale." }}
    </section>
    
    {{ include "./partials/cta.html" title="Not sure?" text="Contact our sales team for a custom quote." button_text="Contact Sales" button_link="/contact" }}
    {{ include "./partials/footer.html" }}
{{ endblock }}
{{ endlayout }}`

    contactPageHTML := `{{ layout "./layouts/guest_layout.html" }}
{{ block "title" }}Contact - This Page{{ endblock }}
{{ block "main" }}
    {{ include "./partials/page_header.html" title="Get in Touch" subtitle="We'd love to hear from you." }}
    
    <div class="max-w-2xl mx-auto py-12">
        <form class="space-y-6">
            <div>
                <label class="block text-sm font-medium text-neutral-400 mb-2">Name</label>
                <input type="text" class="w-full bg-neutral-900 border border-neutral-800 rounded p-3 text-white focus:border-blue-500 focus:outline-none">
            </div>
            <div>
                <label class="block text-sm font-medium text-neutral-400 mb-2">Email</label>
                <input type="email" class="w-full bg-neutral-900 border border-neutral-800 rounded p-3 text-white focus:border-blue-500 focus:outline-none">
            </div>
            <div>
                <label class="block text-sm font-medium text-neutral-400 mb-2">Message</label>
                <textarea rows="4" class="w-full bg-neutral-900 border border-neutral-800 rounded p-3 text-white focus:border-blue-500 focus:outline-none"></textarea>
            </div>
            <button type="submit" class="bg-white text-black font-bold py-3 px-8 rounded hover:bg-neutral-200 transition-colors">Send Message</button>
        </form>
    </div>

    {{ include "./partials/footer.html" }}
{{ endblock }}
{{ endlayout }}`

	defaultNavigationHTML := `<nav class="flex gap-6 border-b border-neutral-800 pb-6 w-full mb-8 items-center">
  <a href='/' class="font-bold text-white text-lg mr-auto">ThisPage</a>
  <a href='/' class="text-xs uppercase tracking-widest hover:text-white transition-colors text-neutral-400">Home</a>
  <a href='/about' class="text-xs uppercase tracking-widest hover:text-white transition-colors text-neutral-400">About</a>
  <a href='/pricing' class="text-xs uppercase tracking-widest hover:text-white transition-colors text-neutral-400">Pricing</a>
  <a href='/contact' class="text-xs uppercase tracking-widest hover:text-white transition-colors text-neutral-400">Contact</a>
</nav>`

    defaultHeroHTML := `<header class="py-20 text-center thispage-block">
    <h1 class="text-5xl font-bold text-white tracking-tight mb-6">Welcome to Your Site</h1>
    <p class="text-xl text-neutral-400 max-w-2xl mx-auto">This is a hero section. It's a great place to introduce your brand or project.</p>
</header>`

    defaultContentHTML := `<section class="py-12 thispage-block">
    <h2 class="text-2xl font-bold text-white mb-4">Content Section</h2>
    <p class="text-neutral-400 leading-relaxed">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
        Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
    </p>
</section>`

    defaultFooterHTML := `<footer class="border-t border-neutral-800 py-8 mt-12 text-center text-sm text-neutral-600 thispage-block">
    &copy; 2024 Your Company. All rights reserved.
</footer>`

    defaultCardHTML := `<div class="bg-neutral-900 border border-neutral-800 rounded-lg overflow-hidden hover:border-neutral-700 transition-colors thispage-block group">
  <div class="h-48 bg-neutral-800 relative overflow-hidden">
      <img src="{{ prop "image" }}" alt="{{ prop "title" }}" class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110">
  </div>
  <div class="p-6">
    <h3 class="text-xl font-bold text-white mb-2">{{ prop "title" }}</h3>
    <p class="text-neutral-400 text-sm mb-4 leading-relaxed">{{ prop "description" }}</p>
    <a href="{{ prop "link" }}" class="text-blue-500 hover:text-blue-400 text-xs font-bold uppercase tracking-widest inline-flex items-center gap-1">
        Read More 
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3"></path></svg>
    </a>
  </div>
</div>`

    defaultCtaHTML := `<section class="py-24 text-center bg-gradient-to-b from-neutral-900 to-black border-y border-neutral-800 thispage-block">
    <div class="max-w-3xl mx-auto px-6">
        <h2 class="text-4xl font-bold text-white mb-6 tracking-tight">{{ prop "title" }}</h2>
        <p class="text-lg text-neutral-400 mb-10 leading-relaxed">{{ prop "text" }}</p>
        <a href="{{ prop "button_link" }}" class="bg-white text-black font-bold py-4 px-10 rounded-full transition-all hover:scale-105 hover:bg-neutral-200 inline-block shadow-lg shadow-white/10">
            {{ prop "button_text" }}
        </a>
    </div>
</section>`

    defaultTestimonialHTML := `<div class="bg-neutral-900/50 p-8 border border-neutral-800 rounded-2xl thispage-block backdrop-blur-sm">
  <div class="flex gap-1 text-blue-500 mb-6">
      <svg class="w-4 h-4 fill-current" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/></svg>
      <svg class="w-4 h-4 fill-current" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/></svg>
      <svg class="w-4 h-4 fill-current" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/></svg>
      <svg class="w-4 h-4 fill-current" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/></svg>
      <svg class="w-4 h-4 fill-current" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/></svg>
  </div>
  <p class="text-xl text-neutral-300 italic mb-8 leading-relaxed font-serif">"{{ prop "quote" }}"</p>
  <div class="flex items-center gap-4 border-t border-neutral-800 pt-6">
    <div class="w-12 h-12 rounded-full bg-gradient-to-tr from-blue-600 to-purple-600 flex items-center justify-center text-white font-bold shadow-lg">
        {{ prop "initials" }}
    </div>
    <div>
        <h4 class="text-white font-bold text-sm">{{ prop "author" }}</h4>
        <p class="text-neutral-500 text-[10px] uppercase tracking-widest font-bold">{{ prop "role" }}</p>
    </div>
  </div>
</div>`

    defaultPageHeaderHTML := `<div class="py-16 border-b border-neutral-800 mb-12 thispage-block">
    <h1 class="text-5xl font-bold text-white mb-4 tracking-tight">{{ prop "title" }}</h1>
    <p class="text-xl text-neutral-400 max-w-2xl">{{ prop "subtitle" }}</p>
</div>`

    defaultFeatureItemHTML := `<div class="p-6 border border-neutral-800 rounded-lg thispage-block hover:bg-neutral-900 transition-colors">
    <div class="w-12 h-12 bg-blue-900/30 text-blue-500 rounded-full flex items-center justify-center mb-4 text-2xl">
        {{ prop "icon" }}
    </div>
    <h3 class="text-lg font-bold text-white mb-2">{{ prop "title" }}</h3>
    <p class="text-neutral-400 text-sm leading-relaxed">{{ prop "description" }}</p>
</div>`

    defaultPricingCardHTML := `<div class="border border-neutral-800 p-8 rounded-2xl bg-neutral-900/20 thispage-block hover:border-blue-600 transition-colors relative flex flex-col">
    <h3 class="text-lg font-medium text-neutral-400 mb-4">{{ prop "plan" }}</h3>
    <div class="text-4xl font-bold text-white mb-6">${{ prop "price" }}<span class="text-sm text-neutral-500 font-normal">/mo</span></div>
    <p class="text-sm text-neutral-400 mb-8 border-b border-neutral-800 pb-8 flex-grow">{{ prop "description" }}</p>
    <a href="#" class="block w-full bg-white text-black font-bold text-center py-3 rounded hover:bg-neutral-200 transition-colors mt-auto">Choose Plan</a>
</div>`

    defaultTeamMemberHTML := `<div class="text-center thispage-block group">
    <div class="w-32 h-32 mx-auto rounded-full overflow-hidden mb-4 border-2 border-neutral-800 group-hover:border-blue-500 transition-colors">
        <img src="{{ prop "image" }}" alt="{{ prop "name" }}" class="w-full h-full object-cover">
    </div>
    <h3 class="text-white font-bold text-lg">{{ prop "name" }}</h3>
    <p class="text-blue-500 text-xs uppercase tracking-widest mt-1">{{ prop "role" }}</p>
</div>`

	filesToCreate := map[string]string{
		filepath.Join(templatesDirPath, "index.html"):         defaultIndexHTML,
		filepath.Join(templatesDirPath, "about.html"):         aboutPageHTML,
		filepath.Join(templatesDirPath, "pricing.html"):       pricingPageHTML,
		filepath.Join(templatesDirPath, "contact.html"):       contactPageHTML,
        filepath.Join(layoutsDirPath, "guest_layout.html"):    guestLayoutHTML,
		filepath.Join(partialsDirPath, "navigation.html"):     defaultNavigationHTML,
		filepath.Join(partialsDirPath, "hero.html"):           defaultHeroHTML,
		filepath.Join(partialsDirPath, "content.html"):        defaultContentHTML,
		filepath.Join(partialsDirPath, "footer.html"):          defaultFooterHTML,
        filepath.Join(partialsDirPath, "card.html"):           defaultCardHTML,
        filepath.Join(partialsDirPath, "cta.html"):            defaultCtaHTML,
        filepath.Join(partialsDirPath, "testimonial.html"):    defaultTestimonialHTML,
        filepath.Join(partialsDirPath, "page_header.html"):    defaultPageHeaderHTML,
        filepath.Join(partialsDirPath, "feature_item.html"):   defaultFeatureItemHTML,
        filepath.Join(partialsDirPath, "pricing_card.html"):   defaultPricingCardHTML,
        filepath.Join(partialsDirPath, "team_member.html"):    defaultTeamMemberHTML,
		filepath.Join(name, "static/input.css"): "@import \"tailwindcss\";\n",
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create default file '%s': %w", filepath.Base(path), err)
		}
	}

	return nil
}
